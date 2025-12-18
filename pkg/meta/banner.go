package meta

import "strings"

const bannerString = `
███╗   ██╗██╗   ██╗ █████╗  █████╗  ██████╗  ██████╗ 
████╗  ██║╚██╗ ██╔╝██╔══██╗██╔══██╗██╔════╝ ██╔═══██╗
██╔██╗ ██║ ╚████╔╝ ███████║███████║██║  ███╗██║   ██║
██║╚██╗██║  ╚██╔╝  ██╔══██║██╔══██║██║   ██║██║   ██║
██║ ╚████║   ██║   ██║  ██║██║  ██║╚██████╔╝╚██████╔╝
╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝  ╚═════╝  
`

const bannerWidth = 53

type banner struct{}

func (b banner) Lines() []string {
	return strings.Split(bannerString, "\n")
}

func (b banner) Width() int {
	return bannerWidth
}

func (b banner) Config() fieldConfig {
	return fieldConfig{
		Alignment: fieldAlignCenter,
	}
}

func getBanner() field {
	return banner{}
}
