package main

import (
	"fmt"

	. "jakobsachs.blog/kvStore/shared"
)

type Client struct {
  
}

func main() {
	req := Request{
		Id: 0,
	}

	fmt.Printf("%#v", req);
}
