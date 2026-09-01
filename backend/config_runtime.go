package backend

import "os"

func init() {
	lookupEnv = os.Getenv
}
