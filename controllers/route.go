package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

// 腾讯地图API配置
const (
	TencentMapKey = "4JIBZ-3WGW7-K66X2-PGOUY-FKDSH-TDF4A"
	TencentAPIURL = "https://apis.map.qq.com/ws/direction/v1/walking/"
)

// 请求参数结构体
type RouteRequest struct {
	StartLat float64 `json:"startLat" binding:"required"`
	StartLng float64 `json:"startLng" binding:"required"`
	EndLat   float64 `json:"endLat" binding:"required"`
	EndLng   float64 `json:"endLng" binding:"required"`
}

// 路径点结构体
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// 响应数据结构体
type RouteResponse struct {
	Distance int     `json:"distance"`
	Duration int     `json:"duration"`
	Path     []Point `json:"path"`
}

// 腾讯地图API响应结构体（修正版）
type TencentMapResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Routes []struct {
			Distance int       `json:"distance"`
			Duration int       `json:"duration"`
			Polyline []float64 `json:"polyline"` // ⚠️ 修正：polyline是数字数组，在routes级别
			Steps    []struct {
				Instruction string `json:"instruction"`
				PolylineIdx []int  `json:"polyline_idx"` // polyline_idx指向polyline数组的索引
				RoadName    string `json:"road_name"`
				Distance    int    `json:"distance"`
			} `json:"steps"`
		} `json:"routes"`
	} `json:"result"`
}

// GetRoute 路线规划处理函数
func GetRoute(c *gin.Context) {
	var req RouteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数不完整:  "+err.Error())
		return
	}

	fmt.Printf("[路线规划] 收到请求 - 起点: (%.6f, %.6f), 终点: (%.6f, %.6f)\n",
		req.StartLat, req.StartLng, req.EndLat, req.EndLng)

	routeResp, err := callTencentMapAPI(req.StartLat, req.StartLng, req.EndLat, req.EndLng)
	if err != nil {
		fmt.Printf("[路线规划] 错误: %v\n", err)
		utils.InternalServerError(c, "路线规划失败", err)
		return
	}

	fmt.Printf("[路线规划] 成功 - 距离: %dm, 时长:  %d分钟, 路径点: %d个\n",
		routeResp.Distance, routeResp.Duration, len(routeResp.Path))

	utils.Success(c, routeResp)
}

// 调用腾讯地图API
func callTencentMapAPI(startLat, startLng, endLat, endLng float64) (*RouteResponse, error) {
	// ⚠️ 修改：腾讯地图API的坐标顺序是 纬度,经度
	url := fmt.Sprintf("%s?from=%.6f,%.6f&to=%.6f,%.6f&key=%s",
		TencentAPIURL, startLat, startLng, endLat, endLng, TencentMapKey)

	fmt.Printf("[腾讯地图API] 请求URL: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求腾讯地图API失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	fmt.Printf("[腾讯地图API] 响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("[腾讯地图API] 响应内容: %s\n", string(body))

	var apiResp TencentMapResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败:  %v", err)
	}

	if apiResp.Status != 0 {
		return nil, fmt.Errorf("腾讯地图API错误[%d]: %s", apiResp.Status, apiResp.Message)
	}

	if len(apiResp.Result.Routes) == 0 {
		return nil, fmt.Errorf("未找到路线")
	}

	route := apiResp.Result.Routes[0]

	fmt.Printf("[路径解析] 路线信息 - 距离: %dm, 时长: %ds, Polyline数组长度: %d\n",
		route.Distance, route.Duration, len(route.Polyline))
	fmt.Printf("[路径解析] 输入起点: (%.6f, %.6f)\n", startLat, startLng)
	fmt.Printf("[路径解析] 输入终点: (%.6f, %.6f)\n", endLat, endLng)

	// ⚠️ 修改：解析腾讯地图的增量编码polyline
	path := decodePolylineFromArray(route.Polyline)

	fmt.Printf("[路径解析] 解析出路径点数: %d\n", len(path))
	if len(path) > 0 {
		fmt.Printf("[路径解析] 路径起点: (%.6f, %.6f)\n", path[0].Lat, path[0].Lng)
		fmt.Printf("[路径解析] 路径终点: (%.6f, %.6f)\n", path[len(path)-1].Lat, path[len(path)-1].Lng)
		fmt.Printf("[路径解析] 起点偏差: (%.6f, %.6f)\n", path[0].Lat-startLat, path[0].Lng-startLng)
		fmt.Printf("[路径解析] 终点偏差: (%.6f, %.6f)\n", path[len(path)-1].Lat-endLat, path[len(path)-1].Lng-endLng)
	}

	// ⚠️ 添加：如果没有路径点，至少返回起点和终点
	if len(path) == 0 {
		fmt.Printf("[路径解析] 警告：未能解析出路径点，返回起终点\n")
		path = []Point{
			{Lat: startLat, Lng: startLng},
			{Lat: endLat, Lng: endLng},
		}
	}

	routeResp := &RouteResponse{
		Distance: route.Distance,
		Duration: (route.Duration + 59) / 60,
		Path:     path,
	}

	return routeResp, nil
}

// ⚠️ 新增：解码腾讯地图的增量编码polyline数组
// 腾讯地图polyline格式：[起始纬度, 起始经度, 纬度增量1, 经度增量1, 纬度增量2, 经度增量2, ...]
// 增量值需要除以100000转换为实际坐标
func decodePolylineFromArray(encoded []float64) []Point {
	points := make([]Point, 0)

	if len(encoded) < 2 {
		return points
	}

	fmt.Printf("[Polyline解码] 输入数组长度: %d\n", len(encoded))
	fmt.Printf("[Polyline解码] 前10个值: %v\n", encoded[:min(10, len(encoded))])

	// 第一个点是绝对坐标
	lat := encoded[0]
	lng := encoded[1]
	points = append(points, Point{Lat: lat, Lng: lng})
	fmt.Printf("[Polyline解码] 第1个点(绝对): (%.6f, %.6f)\n", lat, lng)

	// 后续是增量坐标
	pointIndex := 2
	for i := 2; i < len(encoded); i += 2 {
		if i+1 >= len(encoded) {
			break
		}

		latDelta := encoded[i]
		lngDelta := encoded[i+1]

		// ⚠️ 修改：增量值需要除以1000000（不是100000）
		lat += latDelta / 1000000.0
		lng += lngDelta / 1000000.0

		points = append(points, Point{Lat: lat, Lng: lng})
		pointIndex++
		fmt.Printf("[Polyline解码] 第%d个点(增量): delta=(%.0f, %.0f) -> (%.6f, %.6f)\n",
			pointIndex, latDelta, lngDelta, lat, lng)
	}

	fmt.Printf("[Polyline解码] 总共解析出 %d 个点\n", len(points))
	return points
}

// 辅助函数：返回两个整数中的较小值
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
