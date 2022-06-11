package must

func String(str string, err error) string {
	if err != nil {
		panic(err)
	}

	return str
}
