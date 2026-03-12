package go_zero

import "testing"

func TestBuildBulkInsertQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		rows    int
		want    string
		wantErr bool
	}{
		{
			name:  "zero row keeps original",
			query: "INSERT INTO `t` (`a`,`b`) VALUES (?,?)",
			rows:  0,
			want:  "INSERT INTO `t` (`a`,`b`) VALUES (?,?)",
		},
		{
			name:  "single row keeps original",
			query: "INSERT INTO `t` (`a`,`b`) VALUES (?,?)",
			rows:  1,
			want:  "INSERT INTO `t` (`a`,`b`) VALUES (?,?)",
		},
		{
			name:  "expand multi rows",
			query: "INSERT INTO `t` (`a`,`b`) VALUES (?,?)",
			rows:  3,
			want:  "INSERT INTO `t` (`a`,`b`) VALUES (?,?),(?,?),(?,?)",
		},
		{
			name:  "expand with suffix",
			query: "INSERT INTO `t` (`a`,`b`) VALUES (?,?) ON DUPLICATE KEY UPDATE `a`=VALUES(`a`)",
			rows:  2,
			want:  "INSERT INTO `t` (`a`,`b`) VALUES (?,?),(?,?) ON DUPLICATE KEY UPDATE `a`=VALUES(`a`)",
		},
		{
			name:    "invalid query",
			query:   "INSERT INTO `t` (`a`,`b`)",
			rows:    2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildBulkInsertQuery(tt.query, tt.rows)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
