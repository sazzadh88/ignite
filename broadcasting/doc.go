// Package broadcasting provides event broadcasting capabilities for Ignite.
//
// Broadcasting allows you to send real-time events to connected clients through
// various channels (public, private, presence, encrypted). This package provides
// a Laravel-inspired API with zero external dependencies.
//
// # Basic Usage
//
// Broadcasting an event using the default manager:
//
//	broadcasting.Broadcast.Broadcast(
//		[]broadcasting.Channel{broadcasting.PublicChannel("news")},
//		"NewsPublished",
//		map[string]any{"title": "Breaking News", "content": "..."},
//	)
//
// # Custom Broadcasters
//
// Register a custom broadcaster driver:
//
//	manager := broadcasting.NewManager()
//	manager.Extend("log", broadcasting.NewLogBroadcaster(os.Stdout))
//	manager.SetDefaultDriver("log")
//
// # Broadcastable Events
//
// Implement the ShouldBroadcast interface for automatic broadcasting:
//
//	type OrderShipped struct {
//		OrderID int
//	}
//
//	func (e *OrderShipped) BroadcastOn() []broadcasting.Channel {
//		return []broadcasting.Channel{
//			broadcasting.PrivateChannel(fmt.Sprintf("order.%d", e.OrderID)),
//		}
//	}
//
//	func (e *OrderShipped) BroadcastAs() string {
//		return "OrderShipped"
//	}
//
//	func (e *OrderShipped) BroadcastWith() map[string]any {
//		return map[string]any{"order_id": e.OrderID}
//	}
//
// Then broadcast the event:
//
//	event := &OrderShipped{OrderID: 123}
//	broadcasting.Broadcast.Event(event)
//
// # Channel Authorization
//
// Register authorization callbacks for private and presence channels:
//
//	broadcasting.RegisterChannel("chat.{id}", func(user any, params map[string]string) any {
//		chatID := params["id"]
//		// Check if user can access this chat
//		if canAccessChat(user, chatID) {
//			return map[string]any{"user_id": getUserID(user)}
//		}
//		return nil
//	})
//
// Use the auth handler in your HTTP routes:
//
//	http.HandleFunc("/broadcasting/auth", broadcasting.AuthHandler(func(r *http.Request) any {
//		// Extract authenticated user from request
//		return getAuthenticatedUser(r)
//	}))
//
// # Testing
//
// Use the fake broadcaster for testing:
//
//	fake := broadcasting.Fake()
//	broadcasting.Broadcast.Extend("fake", fake)
//	broadcasting.Broadcast.SetDefaultDriver("fake")
//
//	// Perform actions that trigger broadcasts
//	someFunction()
//
//	// Assert broadcasts
//	if !fake.AssertBroadcast("OrderShipped") {
//		t.Error("OrderShipped event was not broadcast")
//	}
//
//	if !fake.AssertBroadcastOn(broadcasting.PublicChannel("orders"), "OrderShipped") {
//		t.Error("OrderShipped was not broadcast on the correct channel")
//	}
package broadcasting
