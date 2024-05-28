```
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Email string `json:"email"`
}

type Ticket struct {
    ID          int    `json:"id"`
    Summary     string `json:"summary"`
    Description string `json:"description"`
    Submitted  time.Time `json:"submitted"`
    Status     string `json:"status"`
    User        User   `json:"user"` // Embedded User struct
}

```