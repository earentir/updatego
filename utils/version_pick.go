package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func compareSemver(a, b string) int {
	pa := parseSemver(a)
	pb := parseSemver(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(v, ".", 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			continue
		}
		out[i] = n
	}
	return out
}

func findVersionFor(htmlContent, goos, goarch string) (string, error) {
	osName, arch, ext := platformParts(goos, goarch)
	pattern := fmt.Sprintf(`go(\d+\.\d+\.\d+)\.%s-%s%s`,
		regexp.QuoteMeta(osName), regexp.QuoteMeta(arch), regexp.QuoteMeta(ext))
	regex := regexp.MustCompile(pattern)
	all := regex.FindAllStringSubmatch(htmlContent, -1)
	if len(all) == 0 {
		return "", fmt.Errorf("no version found for %s/%s", osName, arch)
	}
	best := all[0][1]
	for _, m := range all[1:] {
		if compareSemver(m[1], best) > 0 {
			best = m[1]
		}
	}
	return best, nil
}
