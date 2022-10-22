package error
type Error struct{
	msg string
	group string
	key string
}
func New(msg string, args...string) *Error{
	group:=""
	key:=""
	if len(args)>0{
		group=args[0]
	}
	if len(args)>1{
		key=args[1]
	}
	return &Error{
		msg : msg,
		group : group,
		key: key,
	}
}

func (err *Error)Msg() string{
	return err.msg
}
func (err *Error)Group() string{
	return err.group
}
func (err *Error)Key() string{
	return err.key
}