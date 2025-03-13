package player

import (
	"fmt"
	"os"

	vlc "github.com/adrg/libvlc-go/v3"
)

type VLCPlayer struct {
	list *MediaList

	player    *vlc.Player
	currMedia *vlc.Media

	next chan struct{}
}

func (p *VLCPlayer) Init() error {
	// Initialize VLC with verbose logging
	err := vlc.Init("--no-video", "--verbose", "2")
	if err != nil {
		return err
	}

	p.next = make(chan struct{})

	p.player, err = vlc.NewPlayer()
	if err != nil {
		return err
	}

	// Disable fullscreen in Docker
	if os.Getenv("DOCKER_ENV") == "true" {
		p.player.SetFullScreen(false)
	} else {
		p.player.SetFullScreen(true)
	}

	manager, err := p.player.EventManager()
	if err != nil {
		return err
	}

	eventCallback := func(event vlc.Event, userData interface{}) {
		if event == vlc.MediaPlayerEndReached {
			p.next <- struct{}{}
		}
	}

	_, err = manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		return err
	}

	go func() {
		for range p.next {
			err := p.PlayNext()
			if err != nil {
				fmt.Println("Error playing next:", err)
			}
		}
	}()

	return nil
}

func (p *VLCPlayer) Shutdown() error {
	if p.player != nil {
		p.player.Stop()
		p.player.Release()
	}
	if p.currMedia != nil {
		p.currMedia.Release()
	}
	return vlc.Release()
}

func (p *VLCPlayer) Play(list *MediaList) error {
	if p.player == nil {
		return ErrPlayerNotInitialized
	}

	p.list = list
	isDocker := os.Getenv("DOCKER_ENV") == "true"

	var err error
	p.currMedia, err = vlc.NewMediaFromPath(p.list.Current())
	if err != nil {
		return err
	}

	// If running in Docker, ensure HLS options are correctly applied
	if isDocker {
		// Ensure HLS output directory exists
		outputDir := "./static" // Change this if needed
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return err
		}

		// Corrected HLS stream options
		hlsOptions := []string{
			"sout=#http{mux=ts,dst=:8080/stream}", // Stream via HTTP
			"sout-keep",                           // Keep the stream alive
		}

		// Apply options to the new media
		for _, opt := range hlsOptions {
			err = p.currMedia.AddOptions(opt)
			if err != nil {
				return err
			}
		}

		fmt.Println("Streaming HLS at: http://localhost:8080/static/stream.m3u8")
	}

	// Set the media for the player
	err = p.player.SetMedia(p.currMedia)
	if err != nil {
		return err
	}

	// Start the player
	return p.player.Play()
}

func (p *VLCPlayer) PlayNext() error {
	if p.player == nil {
		return ErrPlayerNotInitialized
	}

	// Stop the player to reset streaming
	err := p.player.Stop()
	if err != nil {
		return err
	}

	// Release current media
	if p.currMedia != nil {
		p.currMedia.Release()
	}

	// Get the next file from the list
	nextFile := p.list.Advance()

	// Create new media instance for the next file
	p.currMedia, err = vlc.NewMediaFromPath(nextFile)
	if err != nil {
		return err
	}

	// If running in Docker, ensure HLS options are correctly applied
	isDocker := os.Getenv("DOCKER_ENV") == "true"
	if isDocker {
		// Corrected HLS stream options
		hlsOptions := []string{
			"sout=#http{mux=ts,dst=:8080/stream}", // Stream via HTTP
			"sout-keep",                           // Keep the stream alive
		}

		// Apply options to the new media
		for _, opt := range hlsOptions {
			err = p.currMedia.AddOptions(opt)
			if err != nil {
				return err
			}
		}
	}

	// Set the media for the player
	err = p.player.SetMedia(p.currMedia)
	if err != nil {
		return err
	}

	// Start the player again
	return p.player.Play()
}

func (p *VLCPlayer) Next() string {
	return p.list.Next()
}

func (p *VLCPlayer) Current() string {
	return p.list.Current()
}
