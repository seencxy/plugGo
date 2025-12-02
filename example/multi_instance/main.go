package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seencxy/plugGo"
	announcementConfig "github.com/seencxy/plugGo/example/announcement/config"
	"github.com/seencxy/plugGo/registry"

	// Import plugin factory (triggers init auto-registration)
	_ "github.com/seencxy/plugGo/example/announcement"
)

func main() {
	fmt.Println("=== PlugGo Multi-Instance Example ===\n")

	// List all registered plugin factories
	factories := registry.GetAllFactories()
	fmt.Printf("Found %d plugin factories:\n", len(factories))
	for name, factory := range factories {
		fmt.Printf("  - %s (v%s)\n", name, factory.Version())
	}
	fmt.Println()

	// Create multiple announcement plugin instances, each monitoring different sources

	// Instance 1: Monitor GitHub
	fmt.Println("Creating instance 1: GitHub Monitor")
	githubConfig := &announcementConfig.Config{
		Name:     "github-monitor",
		Enabled:  true,
		LogLevel: "info",
		Sources: []announcementConfig.Source{
			{
				Name:     "GitHub Releases",
				URL:      "https://api.github.com/repos/golang/go/releases",
				Interval: 60,
			},
		},
	}

	instance1, err := registry.CreateInstance(
		"announcement",
		"github-monitor",
		githubConfig,
		plugGo.NewStandardLogger("github-monitor", plugGo.InfoLevel),
	)
	if err != nil {
		fmt.Printf("Failed to create instance 1: %v\n", err)
		return
	}
	fmt.Printf("  [OK] Instance created: %s\n", instance1.ID())
	if err := instance1.Start(context.Background()); err != nil {
		fmt.Printf("  [FAIL] Start failed: %v\n", err)
		return
	}
	fmt.Println("  [OK] Instance started\n")

	// Instance 2: Monitor tech blogs
	fmt.Println("Creating instance 2: Tech Blog Monitor")
	blogConfig := &announcementConfig.Config{
		Name:     "blog-monitor",
		Enabled:  true,
		LogLevel: "debug",
		Sources: []announcementConfig.Source{
			{
				Name:     "Tech Blog",
				URL:      "https://blog.golang.org/feed.atom",
				Interval: 120,
			},
			{
				Name:     "Dev News",
				URL:      "https://dev.to/feed/golang",
				Interval: 300,
			},
		},
		Filters: announcementConfig.Filters{
			Keywords: []string{"golang", "performance", "optimization"},
		},
	}

	instance2, err := registry.CreateInstance(
		"announcement",
		"blog-monitor",
		blogConfig,
		plugGo.NewStandardLogger("blog-monitor", plugGo.DebugLevel),
	)
	if err != nil {
		fmt.Printf("Failed to create instance 2: %v\n", err)
		return
	}
	fmt.Printf("  [OK] Instance created: %s\n", instance2.ID())
	if err := instance2.Start(context.Background()); err != nil {
		fmt.Printf("  [FAIL] Start failed: %v\n", err)
		return
	}
	fmt.Println("  [OK] Instance started\n")

	// Instance 3: Use default config
	fmt.Println("Creating instance 3: Using default config")
	instance3, err := registry.CreateInstance(
		"announcement",
		"default-monitor",
		nil, // Use default config
		nil, // Use default logger
	)
	if err != nil {
		fmt.Printf("Failed to create instance 3: %v\n", err)
		return
	}
	fmt.Printf("  [OK] Instance created: %s\n", instance3.ID())
	if err := instance3.Start(context.Background()); err != nil {
		fmt.Printf("  [FAIL] Start failed: %v\n", err)
		return
	}
	fmt.Println("  [OK] Instance started\n")

	// Display all running instances
	fmt.Println("=== Running Instances ===")
	allInstances := registry.GetAllInstances()
	fmt.Printf("Total %d instances:\n", len(allInstances))
	for _, inst := range allInstances {
		plugin := inst.Plugin()
		fmt.Printf("  - ID: %s, Type: %s, Version: %s\n",
			plugin.ID(), plugin.PluginType(), plugin.Version())
	}
	fmt.Println()

	// Wait 5 seconds
	fmt.Println("Waiting 5 seconds to observe...")
	time.Sleep(5 * time.Second)

	// Demo config hot reload
	fmt.Println("\n=== Demo Config Hot Reload ===")
	fmt.Println("Updating instance 1 config...")
	newGithubConfig := &announcementConfig.Config{
		Name:     "github-monitor",
		Enabled:  true,
		LogLevel: "debug",
		Sources: []announcementConfig.Source{
			{
				Name:     "GitHub Releases",
				URL:      "https://api.github.com/repos/golang/go/releases",
				Interval: 30, // Shortened interval
			},
			{
				Name:     "GitHub Issues",
				URL:      "https://api.github.com/repos/golang/go/issues",
				Interval: 60,
			},
		},
	}

	if err := instance1.UpdateConfig(newGithubConfig); err != nil {
		fmt.Printf("  [FAIL] Config update failed: %v\n", err)
	} else {
		fmt.Println("  [OK] Config updated, added a new monitor source")
	}
	fmt.Println()

	// Demo query instances by type
	fmt.Println("=== Query Instances by Type ===")
	announcementInstances := registry.GetInstancesByType("announcement")
	fmt.Printf("Instances of type 'announcement': %d\n", len(announcementInstances))
	for _, inst := range announcementInstances {
		fmt.Printf("  - %s\n", inst.ID())
	}
	fmt.Println()

	// Wait for user interrupt
	fmt.Println("Press Ctrl+C to stop all instances...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully shutdown all instances
	fmt.Println("\n=== Shutting Down All Instances ===")
	for _, inst := range allInstances {
		fmt.Printf("Stopping instance: %s\n", inst.ID())
		if err := inst.Stop(context.Background()); err != nil {
			fmt.Printf("  [FAIL] Stop failed: %v\n", err)
		} else {
			fmt.Println("  [OK] Stopped")
		}

		// Remove from registry
		if err := registry.RemoveInstance(inst.ID()); err != nil {
			fmt.Printf("  [FAIL] Remove failed: %v\n", err)
		}
	}

	fmt.Printf("\nStats: %d factories, %d instances\n",
		registry.CountFactories(), registry.CountInstances())
	fmt.Println("\nAll instances stopped. Goodbye!")
}
