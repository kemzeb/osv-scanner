package osvscanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/osv-scanner/internal/testutility"
	"github.com/google/osv-scanner/pkg/config"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/google/osv-scanner/pkg/reporter"
)

func Test_filterResults(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       models.VulnerabilityResults
		want        models.VulnerabilityResults
		numFiltered int
		path        string
	}

	loadTestCase := func(path string) testCase {
		var testCase testCase
		testCase.input = testutility.LoadJSONFixture[models.VulnerabilityResults](t, filepath.Join(path, "input.json"))
		testCase.want = testutility.LoadJSONFixture[models.VulnerabilityResults](t, filepath.Join(path, "want.json"))
		testCase.numFiltered = len(testCase.input.Flatten()) - len(testCase.want.Flatten())
		testCase.path = path

		return testCase
	}
	tests := []struct {
		name     string
		testCase testCase
	}{
		{
			name:     "filter_everything",
			testCase: loadTestCase("fixtures/filter/all/"),
		},
		{
			name:     "filter_nothing",
			testCase: loadTestCase("fixtures/filter/none/"),
		},
		{
			name:     "filter_partially",
			testCase: loadTestCase("fixtures/filter/some/"),
		},
	}
	for _, tt := range tests {
		tt := tt // Reinitialize for t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &reporter.VoidReporter{}
			// ConfigManager looks for osv-scanner.toml in the source path.
			// Sources in the test input should point to files/folders in the text fixture folder for this to work correctly.
			configManager := config.ConfigManager{
				DefaultConfig: config.Config{},
				ConfigMap:     make(map[string]config.Config),
			}
			got := tt.testCase.input
			filtered := filterResults(r, &got, &configManager, false)
			if diff := cmp.Diff(tt.testCase.want, got); diff != "" {
				out := filepath.Join(tt.testCase.path, "out.json")
				t.Errorf("filterResults() returned an unexpected results (-want, +got):\n%s\n"+
					"Full json output written to %s", diff, out)
				//nolint:staticcheck
				testutility.CreateJSONFixture(t, out, got)
			}
			if filtered != tt.testCase.numFiltered {
				t.Errorf("filterResults() = %v, want %v", filtered, tt.testCase.numFiltered)
			}
		})
	}
}

func Test_scanGit(t *testing.T) {
	t.Parallel()

	type args struct {
		r       reporter.Reporter
		repoDir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantPkg []scannedPackage
	}{
		{
			name: "Example Git repo",
			args: args{
				r:       &reporter.VoidReporter{},
				repoDir: "fixtures/example-git",
			},
			wantErr: false,
			wantPkg: []scannedPackage{
				{
					Commit: "862ac4bd2703b622e85f29f55a2fd8cd6caf8182",
					Source: models.SourceInfo{
						Path: "fixtures/example-git",
						Type: "git",
					},
				},
			},
		},
	}

	err := os.Rename("fixtures/example-git/git-hidden", "fixtures/example-git/.git")
	if err != nil {
		t.Errorf("can't find git-hidden folder")
	}

	for _, tt := range tests {
		pkg, err := scanGit(tt.args.r, tt.args.repoDir)
		if (err != nil) != tt.wantErr {
			t.Errorf("scanGit() error = %v, wantErr %v", err, tt.wantErr)
		}
		if diff := cmp.Diff(tt.wantPkg, pkg); diff != "" {
			t.Errorf("scanGit() package = %v, wantPackage %v", pkg, tt.wantPkg)
		}
	}

	err = os.Rename("fixtures/example-git/.git", "fixtures/example-git/git-hidden")
	if err != nil {
		t.Errorf("can't find .git folder")
	}
}
