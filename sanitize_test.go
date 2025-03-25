package main

import (
	"testing"
)

func Test_sanitizeSql(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"numeric", args{"1234567890"}, "1234567890", false},
		{"alphanum", args{"1234567890abcdEFG"}, "1234567890abcdEFG", false},
		{"email", args{"test-here_1@nowhere.com"}, "test-here_1@nowhere.com", false},
		{"file ", args{"../file"}, "", true},
		{"space ", args{"; 1 OR 1"}, "", true},
		{"backslash ", args{"\\0"}, "", true},
		{"#", args{"#metoo"}, "", true},
		// From chatGPT3
		{"drop table", args{"'; DROP TABLE users; --"}, "", true},
		{"OR 1", args{"' OR 1=1; --"}, "", true},
		{"UNION", args{"' UNION SELECT * FROM users; --"}, "", true},
		{"a", args{"' OR 'a'='a"}, "", true},
		{"insert", args{"'; INSERT INTO users (username, password) VALUES ('hacker', 'password'); --"}, "", true},
		{"sql insert", args{"'; INSERT INTO users (username, password) VALUES ('hacker', 'password'); --"}, "", true},
		{"sql update", args{"'; UPDATE users SET password='hacked' WHERE username='admin'; --"}, "", true},
		{"sql delete", args{"'; DELETE FROM users WHERE username='admin'; --"}, "", true},
		{"sql sleep", args{"' OR SLEEP(5) --"}, "", true},
		{"sql load_file", args{"' OR LOAD_FILE('/etc/passwd') --"}, "", true},
		{"sql into outfile", args{"' INTO OUTFILE '/tmp/hacked.txt' --"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeSql(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("sanitizeSql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sanitizeSql() = %v, want %v", got, tt.want)
			}
		})
	}
}
