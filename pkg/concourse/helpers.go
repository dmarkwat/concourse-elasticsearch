package concourse

func MapVersion(vs []string, f func(string) Version) []Version {
	vsm := make([]Version, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
