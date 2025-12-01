package main

import (
	"context"
	"fmt"

	"github.com/seencxy/plugGo"

	// Import all plugins (triggers init to auto-register Entry registration functions)
	_ "github.com/seencxy/plugGo/example/announcement"
)

func main() {
	fmt.Println("=== PlugGo Framework Example - Multi-Instance Demo ===")
	fmt.Println()

	// Create Boot bootstrapper
	// Reads boot.yaml from current directory by default
	boot := plugGo.NewBoot()

	// Optional: add before bootstrap hook
	boot.AddHookFuncBeforeBootstrap("AnnouncementEntry", "official", func(ctx context.Context) {
		fmt.Println("[Hook] Before bootstrapping official announcement entry")
	})

	// Optional: add after bootstrap hook
	boot.AddHookFuncAfterBootstrap("AnnouncementEntry", "official", func(ctx context.Context) {
		fmt.Println("[Hook] After bootstrapping official announcement entry")
	})

	// Optional: add shutdown hook
	boot.AddShutdownHookFunc("cleanup", func() {
		fmt.Println("[Hook] Running cleanup tasks...")
	})

	// Bootstrap all Entries
	fmt.Println("Bootstrapping all entries...")
	boot.Bootstrap(context.Background())

	// ===== Multi-instance demo =====
	fmt.Println()
	fmt.Println("=== Multi-Instance Info ===")

	// Count Entries
	fmt.Printf("Total entries: %d\n", boot.CountEntries())

	// Get all instances of specified type
	announcementEntries := boot.GetEntriesByType("AnnouncementEntry")
	fmt.Printf("AnnouncementEntry instances: %d\n", len(announcementEntries))
	for name, entry := range announcementEntries {
		fmt.Printf("  - %s: %s\n", name, entry.GetDescription())
	}

	// Get single instance
	if official := boot.GetEntry("AnnouncementEntry", "official"); official != nil {
		fmt.Printf("\nDirect access to 'official': %s\n", official.String())
	}

	fmt.Println()
	fmt.Println("=================================")
	fmt.Println("All entries bootstrapped!")
	fmt.Println("Press Ctrl+C to shutdown...")
	fmt.Println("=================================")
	fmt.Println()

	// Note: To monitor plugin status, you can access the plugin instance
	// through your specific Entry implementation and call Status() or
	// subscribe to StatusNotify() channel. See template/entry.go for example.

	// Wait for shutdown signal and gracefully exit
	boot.WaitForShutdownSig(context.Background())

	fmt.Println()
	fmt.Println("All entries stopped. Goodbye!")
}
