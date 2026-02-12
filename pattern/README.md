# 🎨 Design Patterns in Go

> Master the ***Gang of Four*** design patterns through ***practical Go implementations***. Each pattern solves real-world problems with clean, reusable code.

---

## 📚 What are Design Patterns?

**Design patterns** are ***proven solutions to common problems*** in object-oriented software design. They provide:

- ✅ **Reusable solutions** — Tested approaches to recurring design problems
- ✅ **Communication** — Common vocabulary for developers
- ✅ **Best practices** — Encapsulate design knowledge and experience
- ✅ **Flexibility** — Make code more adaptable to change

> *"Design Patterns: Elements of Reusable Object-Oriented Software"* — Gang of Four (Gamma, Helm, Johnson, Vlissides)

---

## 🏗️ Pattern Categories

### 1️⃣ **Behavioral Patterns** — How objects interact & distribute responsibility

| Pattern | Purpose | Real-World Example |
|:---|:---|:---|
| **Chain of Responsibility** | Pass request along chain of handlers | Medical clinic triage system |
| **Command** | Encapsulate request as object | TV remote control commands |
| **Iterator** | Access collection sequentially | Traversing user lists |
| **Mediator** | Centralize object communication | Train station manager |
| **Memento** | Capture & restore state | Undo/Redo functionality |
| **Observer** | Notify multiple objects of state change | Price change notifications |

**Get started:** See detailed explanations below ⬇️

---

## 🔍 Behavioral Patterns Deep Dive

### 🏥 **Chain of Responsibility** (`behavioral/chain_of_responsibility/`)

**Problem:** How to handle a request through a ***sequence of handlers*** without knowing which handler will process it?

**Solution:** Create a chain where each handler decides to process or pass to next.

**Real Example:** Hospital admission
```
Patient → Reception (check-in) → Doctor (diagnosis) → Cashier (payment)
```

**Key Files:**
- `patient.go` — Request object
- `reception.go`, `doctor.go`, `cashier.go` — Handlers in chain
- `medical.go` — Orchestrates the flow

**When to use:**
- ✅ Multiple handlers for single request
- ✅ Handler unknown at compile time
- ✅ Flexible request processing

---

### 📱 **Command** (`behavioral/command/`)

**Problem:** How to ***encapsulate a request as an object*** so clients can parameterize it?

**Solution:** Represent action as a `Command` object with `Execute()` method.

**Real Example:** TV remote control
```
Button.Press() → TurnOnCommand.Execute() → TV.On()
```

**Key Files:**
- `command.go` — Command interface
- `on_command.go`, `off_command.go` — Concrete commands
- `button.go` — Invoker (executes commands)
- `tv.go` — Receiver (performs actual action)

**When to use:**
- ✅ Parameterize objects with operations
- ✅ Queue, log, or undo requests
- ✅ Decouple sender from receiver

---

### 📋 **Iterator** (`behavioral/iterator/`)

**Problem:** How to ***access collection elements sequentially*** without exposing underlying structure?

**Solution:** Create iterator that encapsulates traversal logic.

**Real Example:** User collection iteration
```
for user in userIterator.Next() {
    process(user)
}
```

**Key Files:**
- `iterator.go` — Iterator interface
- `user_iterator.go` — Concrete iterator implementation
- `user_collection.go` — Collection with iterator
- `user.go` — Element object

**When to use:**
- ✅ Access collections uniformly
- ✅ Hide internal structure
- ✅ Support multiple traversals

---

### 🚂 **Mediator** (`behavioral/mediator/`)

**Problem:** How to ***reduce coupling*** when objects need to communicate extensively?

**Solution:** Introduce mediator object that encapsulates communication.

**Real Example:** Train station coordination
```
Trains ←→ Station Manager ←→ Track Assignment
```

**Key Files:**
- `mediator.go` — Mediator interface
- `station_manager.go` — Concrete mediator
- `train.go`, `passenger_train.go`, `freight_train.go` — Colleagues

**When to use:**
- ✅ Complex object interactions
- ✅ Reduce interdependencies
- ✅ Centralize control logic

---

### 💾 **Memento** (`behavioral/memento/`)

**Problem:** How to ***capture & externalize state*** without violating encapsulation?

**Solution:** Create memento to save object state, restored later via caretaker.

**Real Example:** Undo/Redo functionality
```
editor.Save() → Memento(state) → Caretaker.Store() → Undo() → Restore(Memento)
```

**Key Files:**
- `memento.go` — Snapshot of state
- `originator.go` — Object whose state is saved
- `caretaker.go` — Manages mementos

**When to use:**
- ✅ Undo/Redo functionality
- ✅ State restoration
- ✅ Preserve encapsulation

---

### 👁️ **Observer** (`behavioral/observer/`)

**Problem:** How to ***notify multiple objects*** when state changes without coupling?

**Solution:** Define one-to-many dependency so when subject changes, observers auto-update.

**Real Example:** Price change notifications
```
Item.setPrice(100) → notifies all Customer subscribers
```

**Key Files:**
- `item.go` — Subject (observable)
- `customer.go` — Observer (listener)
- Dependencies & state management

**When to use:**
- ✅ Event-driven architectures
- ✅ Real-time notifications
- ✅ Loose coupling requirement

---

## 🚀 How to Explore Patterns

### 1️⃣ **Read Pattern Code**
```bash
cat behavioral/observer/main.go
```

### 2️⃣ **Run the Example**
```bash
cd behavioral/observer
go run main.go
```

### 3️⃣ **Understand the Flow**
- Identify the problem being solved
- See how pattern structures the solution
- Notice how coupling is reduced

### 4️⃣ **Apply to Your Code**
- Recognize pattern usage in your projects
- Refactor problematic code using patterns
- Reuse in future projects

---

## 💡 Key Principles

| Principle | Meaning |
|:---|:---|
| **Open/Closed** | Open for extension, closed for modification |
| **Single Responsibility** | One reason to change |
| **Dependency Inversion** | Depend on abstractions, not concretions |
| **Interface Segregation** | Many specific interfaces vs one general |
| **Composition over Inheritance** | Prefer composition for flexibility |

---

## 🗺️ Pattern Selection Guide

**Need to share data between objects?** → Observer
**Need to decouple sender from receiver?** → Command, Mediator
**Need to traverse collections?** → Iterator
**Need to save/restore state?** → Memento
**Need sequential processing?** → Chain of Responsibility

---

## 📖 Coming Soon

- 🏗️ **Creational Patterns** — Singleton, Factory, Abstract Factory, Builder
- 🧩 **Structural Patterns** — Adapter, Bridge, Composite, Decorator, Facade, Proxy

---

## 🎯 Learning Tips

1. **Master one pattern at a time** — Don't try to learn all at once
2. **Understand the problem first** — Then see how pattern solves it
3. **See real examples** — Run code and modify it
4. **Notice the coupling reduction** — That's the main benefit
5. **Apply patterns wisely** — Avoid over-engineering

---

## 📚 Further Reading

- *Design Patterns: Elements of Reusable Object-Oriented Software* — Gang of Four
- *Head First Design Patterns* — Freeman, Freeman, Bates, Sierra
- [Refactoring Guru Patterns](https://refactoring.guru/design-patterns)

---

**Happy Pattern Learning!** 🚀
