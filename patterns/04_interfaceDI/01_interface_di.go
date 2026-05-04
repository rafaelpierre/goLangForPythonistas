// Topic: Dependency Injection via Interfaces
//
// In Python you might inject dependencies by passing objects, or by using a
// DI framework (injector, dependency_injector). In Go, the idiomatic answer
// is simpler: define a small interface, accept it as a parameter, and let the
// caller decide what to pass. No framework needed.
//
// Go interfaces are satisfied implicitly -- any type that has the required
// methods satisfies the interface, whether or not it knows about it. This is
// Go's structural typing (similar to Python's duck typing, but checked at
// compile time).
//
// Key principle: define interfaces at the point of USE, not the point of
// implementation. A function that needs to store and retrieve data defines a
// 'Store' interface -- it doesn't import the SQL package. The SQL implementation
// lives elsewhere and doesn't know it's being injected.
//
// "Accept interfaces, return structs." -- common Go proverb.
//   - Accept interface: callers pass any implementation (real, fake, mock)
//   - Return struct: callers get a concrete, usable value (not limited by an interface)
//
// Real-world use: swapping a real DB for an in-memory fake in tests; swapping
// HTTP clients for a stub in unit tests; pluggable storage backends (S3, GCS, disk).
//
// Run: go run 01_interface_di.go

//go:build ignore

package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

type Order struct {
	ID        string
	UserID    string
	Amount    float64
	Status    string
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------
// Interfaces defined at the point of USE
//
// Each interface is as small as possible (interface segregation).
// Don't lump everything into one giant Store -- the service only needs what
// it actually calls.
// ---------------------------------------------------------------------------

// OrderStore is what the service needs from persistence.
// It knows nothing about SQL, Redis, or S3.
type OrderStore interface {
	Save(order Order) error
	FindByID(id string) (Order, error)
	FindByUserID(userID string) ([]Order, error)
}

// Notifier sends order confirmation messages.
// The service doesn't care if this is email, SMS, or Slack.
type Notifier interface {
	Notify(userID, message string) error
}

// ---------------------------------------------------------------------------
// Production implementations
// ---------------------------------------------------------------------------

// SQLOrderStore would hit a real database. For this exercise it's fake.
// In real code: import "database/sql" or "github.com/jackc/pgx/v5".
type SQLOrderStore struct {
	// db *sql.DB  -- would be here in real code
	rows map[string]Order // fake in-memory stand-in
}

func NewSQLOrderStore() *SQLOrderStore {
	return &SQLOrderStore{rows: make(map[string]Order)}
}

func (s *SQLOrderStore) Save(order Order) error {
	if _, exists := s.rows[order.ID]; exists {
		return fmt.Errorf("SQLOrderStore.Save: duplicate order id %q", order.ID)
	}
	s.rows[order.ID] = order
	fmt.Printf("[SQL] INSERT INTO orders (id=%s, user=%s, amount=%.2f)\n",
		order.ID, order.UserID, order.Amount)
	return nil
}

func (s *SQLOrderStore) FindByID(id string) (Order, error) {
	order, ok := s.rows[id]
	if !ok {
		return Order{}, fmt.Errorf("SQLOrderStore.FindByID: %w", ErrOrderNotFound)
	}
	return order, nil
}

func (s *SQLOrderStore) FindByUserID(userID string) ([]Order, error) {
	var result []Order
	for _, o := range s.rows {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result, nil
}

// EmailNotifier sends email (fake here -- would use net/smtp or a vendor SDK).
type EmailNotifier struct {
	fromAddress string
}

func NewEmailNotifier(from string) *EmailNotifier {
	return &EmailNotifier{fromAddress: from}
}

func (n *EmailNotifier) Notify(userID, message string) error {
	fmt.Printf("[Email] from=%s to=user:%s body=%q\n", n.fromAddress, userID, message)
	return nil
}

// ---------------------------------------------------------------------------
// Test/fake implementations (would live in _test.go in real code)
//
// Because the interfaces are small, fakes are trivial to write.
// No mocking framework needed.
// ---------------------------------------------------------------------------

// InMemoryOrderStore is a fake used in tests.
type InMemoryOrderStore struct {
	orders map[string]Order
	// SaveCalled is a test probe: lets the test verify Save was called.
	SaveCalled int
}

func NewInMemoryOrderStore() *InMemoryOrderStore {
	return &InMemoryOrderStore{orders: make(map[string]Order)}
}

func (s *InMemoryOrderStore) Save(order Order) error {
	s.SaveCalled++
	s.orders[order.ID] = order
	return nil
}

func (s *InMemoryOrderStore) FindByID(id string) (Order, error) {
	o, ok := s.orders[id]
	if !ok {
		return Order{}, ErrOrderNotFound
	}
	return o, nil
}

func (s *InMemoryOrderStore) FindByUserID(userID string) ([]Order, error) {
	var result []Order
	for _, o := range s.orders {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result, nil
}

// CapturingNotifier records messages instead of sending them.
type CapturingNotifier struct {
	Sent []string // test can inspect this
}

func (n *CapturingNotifier) Notify(userID, message string) error {
	n.Sent = append(n.Sent, fmt.Sprintf("user=%s: %s", userID, message))
	return nil
}

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var ErrOrderNotFound = errors.New("order not found")

// ---------------------------------------------------------------------------
// Service: accepts interfaces, knows nothing about implementations
// ---------------------------------------------------------------------------

// OrderService contains business logic. Its only knowledge of storage and
// notifications comes through the interfaces -- not through concrete types.
type OrderService struct {
	store    OrderStore
	notifier Notifier
}

// NewOrderService is the dependency injection point.
// Callers decide which implementations to wire in.
func NewOrderService(store OrderStore, notifier Notifier) *OrderService {
	return &OrderService{store: store, notifier: notifier}
}

func (svc *OrderService) PlaceOrder(userID string, amount float64) (Order, error) {
	if userID == "" {
		return Order{}, fmt.Errorf("PlaceOrder: userID must not be empty")
	}
	if amount <= 0 {
		return Order{}, fmt.Errorf("PlaceOrder: amount must be positive, got %.2f", amount)
	}

	order := Order{
		ID:        fmt.Sprintf("ord-%d", time.Now().UnixNano()),
		UserID:    userID,
		Amount:    amount,
		Status:    "placed",
		CreatedAt: time.Now(),
	}

	if err := svc.store.Save(order); err != nil {
		return Order{}, fmt.Errorf("PlaceOrder: save: %w", err)
	}

	msg := fmt.Sprintf("Order %s for $%.2f placed successfully.", order.ID, order.Amount)
	if err := svc.notifier.Notify(userID, msg); err != nil {
		// Non-fatal: notification failure shouldn't roll back the order.
		// In real code, you'd emit a metric and maybe retry asynchronously.
		fmt.Printf("PlaceOrder: notification failed (non-fatal): %v\n", err)
	}

	return order, nil
}

func (svc *OrderService) GetOrder(id string) (Order, error) {
	order, err := svc.store.FindByID(id)
	if err != nil {
		return Order{}, fmt.Errorf("GetOrder: %w", err)
	}
	return order, nil
}

func (svc *OrderService) ListUserOrders(userID string) ([]Order, error) {
	orders, err := svc.store.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("ListUserOrders: %w", err)
	}
	return orders, nil
}

func main() {
	fmt.Println("=== Production wiring (SQL store + Email notifier) ===")
	prodSvc := NewOrderService(
		NewSQLOrderStore(),
		NewEmailNotifier("orders@example.com"),
	)

	order, err := prodSvc.PlaceOrder("u1", 99.95)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("placed: %+v\n", order)

	fetched, err := prodSvc.GetOrder(order.ID)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("fetched: id=%s user=%s amount=%.2f\n", fetched.ID, fetched.UserID, fetched.Amount)
	}

	fmt.Println("\n=== Test wiring (in-memory store + capturing notifier) ===")
	// This is exactly what a unit test would do -- no real DB or email server.
	memStore := NewInMemoryOrderStore()
	capturer := &CapturingNotifier{}
	testSvc := NewOrderService(memStore, capturer)

	_, err = testSvc.PlaceOrder("u2", 49.00)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_, err = testSvc.PlaceOrder("u2", 12.50)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("Save called %d times\n", memStore.SaveCalled)
	fmt.Printf("Notifications sent: %d\n", len(capturer.Sent))
	for _, msg := range capturer.Sent {
		fmt.Println(" -", msg)
	}

	orders, _ := testSvc.ListUserOrders("u2")
	total := 0.0
	for _, o := range orders {
		total += o.Amount
	}
	fmt.Printf("u2 has %d orders, total $%.2f\n", len(orders), total)

	fmt.Println("\n=== Error path: order not found ===")
	_, err = testSvc.GetOrder("nonexistent")
	fmt.Println("error:", err)
	fmt.Println("is ErrOrderNotFound?", errors.Is(err, ErrOrderNotFound))

	// TODO 1: Add a CancelOrder(id, reason string) error method to OrderService.
	// It should fetch the order, check that Status == "placed" (return an error
	// if it's already "cancelled"), set Status = "cancelled", save it back with
	// a new Update(order Order) error method on OrderStore, and notify the user.
	// Implement Update on both SQLOrderStore and InMemoryOrderStore.

	// TODO 2: Add a FlakyNotifier that fails every other Notify call (use a
	// counter field). Wire it into testSvc and verify PlaceOrder still
	// succeeds (notification failure is non-fatal), and that the notification
	// count increments only for successful sends.

	// TODO 3: The OrderStore interface returns a single error from Save even
	// on a duplicate ID. Add a new sentinel ErrDuplicateOrder and make
	// SQLOrderStore.Save wrap it on conflict. Then add a PlaceOrderIdempotent
	// method to the service that treats ErrDuplicateOrder as success (fetches
	// and returns the existing order) rather than an error.

	// STRETCH: Move InMemoryOrderStore and CapturingNotifier into a
	// separate file named 'fakes.go' in the same directory. Notice that
	// because both files share `//go:build ignore`, you now need to run them
	// together: `go run 01_interface_di.go fakes.go`. This mirrors how
	// production code splits implementations across files in the same package.
	_ = strings.ToLower // suppress unused import in case TODO 3 is not done yet
}
