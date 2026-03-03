package effect

// Constructor is a function that creates a new Effect instance.
type Constructor func() Effect

var registry = map[string]Constructor{}

// Register adds an effect constructor to the global registry.
func Register(name string, ctor Constructor) {
	registry[name] = ctor
}

// Create instantiates a named effect from the registry.
// Returns nil if the name is not registered.
func Create(name string) Effect {
	ctor, ok := registry[name]
	if !ok {
		return nil
	}
	return ctor()
}

// Names returns all registered effect names.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
