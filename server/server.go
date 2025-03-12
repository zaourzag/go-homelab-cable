package server

import (
	"embed"
	"fmt"
	"math/rand"
	"text/template"
	"time"

	"github.com/clabland/go-homelab-cable/network"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"path/filepath"
)

// embedded files
var (
	//go:embed static/*
	staticFS embed.FS
	//go:embed templates/*.html
	templatesFS embed.FS
)

type Server struct {
	port    string
	Network *network.Network
}

func NewServer(port string, n *network.Network) *Server {
	rand.Seed(time.Now().UnixNano())
	return &Server{
		port:    port,
		Network: n,
	}
}

func (s *Server) Serve() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// need to use MustSubFS since the embedded fs by default includes the
	// subfolder name (in this case "static")
	// if the subfolder name changes, both the //go:embed directive
	// and this will need to be updated
	e.StaticFS("/", echo.MustSubFS(staticFS, "static"))

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseFS(templatesFS, "templates/*.html")),
	}

	e.Renderer = renderer
	e.GET("/api/networks", s.getNetworks)

	e.GET("/api/networks/:callsign/channels", s.getChannels)
	e.GET("/api/networks/:callsign/channels/:channel_id", s.getChannel)
	e.PUT("/api/networks/:callsign/channels/:channel_id/set_live", s.setChannelLive)
	e.PUT("/api/networks/:callsign/channels/:channel_id/play_next", s.playNext)
	e.GET("/api/networks/:callsign/live", s.liveChannel)

	// Routes that always just act upon the current live channel
	e.PUT("/api/networks/:callsign/live/next", s.playLiveNext)

	e.GET("/htmx/meta", s.getHtmxMeta)
	e.GET("/htmx/status", s.getHtmxStatus)
	e.PUT("/htmx/live/next", s.htmxPlayLiveNext)

	// Add a new route to serve the HLS stream
	e.GET("/stream.m3u8", s.serveHLSStream)
	e.GET("/segment-:segment.ts", s.serveHLSSegment)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", s.port)))
}

// Implement a handler function to serve the HLS stream
func (s *Server) serveHLSStream(c echo.Context) error {
	return c.File(filepath.Join("static", "stream.m3u8"))
}

// Implement a handler function to serve the HLS segment
func (s *Server) serveHLSSegment(c echo.Context) error {
	segment := c.Param("segment")
	return c.File(filepath.Join("static", fmt.Sprintf("segment-%s.ts", segment)))
}
