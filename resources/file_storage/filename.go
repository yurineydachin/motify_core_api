package file_storage

import (
	"crypto/md5"
	"fmt"
	"strings"
)

const (
	from   = 5
	level2 = 10
	level3 = 10
)

func GenerateFileName(root, filename string, id uint64) string {
	return fmt.Sprintf(
		"%s/%s/%s.%s",
		root,
		fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%d", root, id))))[from:level2+from],
		fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%d-%s", root, id, filename))))[from:level3+from],
		filename[strings.LastIndex(filename, ".")+1:],
	)
}
