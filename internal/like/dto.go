package like

import "errors"

var (
	errPostNotFound    = errors.New("post not found")
	errCannotLikeOwn   = errors.New("you cannot like your own post")
	errAlreadyLiked    = errors.New("you have already liked this post")
	errLikeNotFound    = errors.New("you have not liked this post")
)