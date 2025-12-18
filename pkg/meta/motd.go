package meta

import (
	"fmt"
	"math/rand"
	"strings"
)

const catTemplate = `          ┌%s┐
 │\_│\    │%s│
 │%c %c \   └%s┘
 │ %c   \ \   
 │______\/   
`

const catTemplateWidth = 12

type cat struct {
	msg      string
	LeftEye  rune
	RightEye rune
	Mouth    rune
}

var cats = []cat{
	{
		"Waiting for something to happen?",
		'.',
		'.',
		'w',
	},
	{
		"Meow?",
		'o',
		'o',
		'w',
	},
	{
		"Puuuuuurrrr...",
		'-',
		'-',
		'w',
	},
}

func (c cat) Lines() []string {
	msgBoxBorder := strings.Repeat("─", len(c.msg))
	return strings.Split(fmt.Sprintf(
		catTemplate,
		msgBoxBorder,
		c.msg,
		c.LeftEye,
		c.RightEye,
		msgBoxBorder,
		c.Mouth,
	),
		"\n",
	)
}

func (c cat) Width() int {
	return catTemplateWidth + len(c.msg)
}

func (c cat) Config() fieldConfig {
	return fieldConfig{
		Alignment: fieldAlignCenter,
	}
}

func getMotd() field {
	return cats[rand.Intn(len(cats))]
}
