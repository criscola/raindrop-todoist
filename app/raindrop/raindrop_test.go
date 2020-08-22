package raindrop

import (
	"fmt"
	"testing"
)

func TestClient_GetBookmarkTags(t *testing.T) {
	c := New("9d8ca8f8-4986-42df-85e8-10a1a2ca5802", "RL", nil)

	err := c.RemovePostponedTagFromBookmark(187833460)
	if err != nil {
		fmt.Println(err)
	}
}