package service

import "time"

// Clock абстрагує час, щоб бізнес-логіку можна було тестувати детерміновано.
type Clock interface{ Now() time.Time }

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }
