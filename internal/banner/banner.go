package banner

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const Version = "0.3.5-INDEV"

const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorDim    = "\033[2m"
)

func PrintBanner() {
	banner := `
    ____          _       ____                           
   / ___| |    __ _| |__   / ___|  ___ _ ____   _____ _ __ 
  | |   | |   / _' | '_ \  \___ \ / _ \ '__\ \ / / _ \ '__|
  | |___| |__| (_| | |_) |  ___) |  __/ |   \ V /  __/ |   
   \____|_____\__,_|_.__/  |____/ \___|_|    \_/ \___|_|   
`
	fmt.Print(colorCyan + banner + colorReset)
	fmt.Printf("  %sVersion %s%s\n\n", colorDim, Version, colorReset)
}

func PrintRoutes(r *gin.Engine) {
	routes := r.Routes()

	groups := make(map[string][]gin.RouteInfo)
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/debug") {
			continue
		}

		group := getRouteGroup(route.Path)
		groups[group] = append(groups[group], route)
	}
	fmt.Println("📡 API Endpoints:")
	fmt.Println(strings.Repeat("─", 60))

	order := []string{"Core", "Auth", "Admin", "Users", "Classrooms", "History"}
	for _, groupName := range order {
		if routes, ok := groups[groupName]; ok {
			printRouteGroup(groupName, routes)
		}
	}
	fmt.Println(strings.Repeat("─", 60))
}

func getRouteGroup(path string) string {
	switch {
	case path == "/health" || path == "/ws" || path == "/compile":
		return "Core"
	case strings.HasPrefix(path, "/login") || path == "/validate":
		return "Auth"
	case strings.HasPrefix(path, "/admin"):
		return "Admin"
	case strings.HasPrefix(path, "/users") || path == "/profile":
		return "Users"
	case strings.HasPrefix(path, "/classrooms"):
		return "Classrooms"
	case strings.HasPrefix(path, "/history"):
		return "History"
	default:
		return "Other"
	}
}

func printRouteGroup(name string, routes []gin.RouteInfo) {
	fmt.Printf("  %s[%s]%s\n", colorYellow, name, colorReset)
	for _, route := range routes {
		if route.Method == "OPTIONS" || route.Method == "HEAD" {
			continue
		}
		methodColor := getMethodColor(route.Method)
		fmt.Printf("    %s%-6s%s %s\n", methodColor, route.Method, colorReset, route.Path)
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return colorGreen
	case "POST":
		return colorCyan
	case "PUT":
		return colorYellow
	case "DELETE":
		return "\033[31m"
	default:
		return colorReset
	}
}

func PrintStartup(port string) {
	fmt.Printf("🚀 %sServer running on port %s%s\n", colorGreen, port, colorReset)
}
