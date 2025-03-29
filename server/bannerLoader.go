package server

import (
	"msgcenter/utils/banner"
	"os"
)

func (s *Server) loadBanner() {
	bannerCfg := s.Consul.GetBanner()
	routes := s.App.GetRoutes()
	routeUrls := make([]string, 0)
	for _, route := range routes {
		routeUrls = append(routeUrls, route.Path+"-"+route.Method+"("+route.Name+")")
	}
	banner.Show(bannerCfg.AppName, bannerCfg.Flag, s.LocalConfig.Ip, os.Getpid(), routeUrls)
}
