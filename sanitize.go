package main

import (
	"fmt"
	"regexp"
)

var (
	validName = regexp.MustCompile(`^[a-zA-Z0-9_@\.\-]+$`)

	loginName  = regexp.MustCompile(`^[a-zA-Z0-9_@\.\-]+$`)
	ipregexp   = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	intregexp  = regexp.MustCompile(`^([0-9]{4})$`)
	mailregexp = regexp.MustCompile(`^[a-zA-Z0-9_\.\-]+@[a-zA-Z0-9_\.\-]+$`)
)

func sanitizeSql(name string) (string, error) {
	// Use a regular expression to match only valid table name characters

	if !validName.MatchString(name) {
		return "", fmt.Errorf("invalid name: %s", name)
	}
	return name, nil
}

func sanitizeByType(name string, match *regexp.Regexp) (string, error) {
	// Use a regular expression to match only valid table name characters
	if !match.MatchString(name) {
		return "", fmt.Errorf("sanitizeByType error with: %s", name)
	}
	return name, nil
}

// https://stackoverflow.com/questions/31647406/mysql-real-escape-string-equivalent-for-golang
//
// func Escape(sql string) string {
//     dest := make([]byte, 0, 2*len(sql))
//     var escape byte
//     for i := 0; i < len(sql); i++ {
//         c := sql[i]

//         escape = 0

//         switch c {
//         case 0: /* Must be escaped for 'mysql' */
//             escape = '0'
//             break
//         case '\n': /* Must be escaped for logs */
//             escape = 'n'
//             break
//         case '\r':
//             escape = 'r'
//             break
//         case '\\':
//             escape = '\\'
//             break
//         case '\'':
//             escape = '\''
//             break
//         case '"': /* Better safe than sorry */
//             escape = '"'
//             break
//         case '\032': //十进制26,八进制32,十六进制1a, /* This gives problems on Win32 */
//             escape = 'Z'
//         }

//         if escape != 0 {
//             dest = append(dest, '\\', escape)
//         } else {
//             dest = append(dest, c)
//         }
//     }

//     return string(dest)
// }
