package events

// Event represents the common interface for all domain events.
//
// An event represents something that took place in the domain. They are always
// named with a past-participle verb, such as "OrderConfirmed". It's not
// unusual, but not required, for an event to name an aggregate or entity that
// it relates to.
//
// Since an event represents something in the past, it can be considered a
// statement of fact and used to take decisions in other parts of the system.
type Event interface{}
