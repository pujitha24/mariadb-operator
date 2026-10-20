package job

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newJob(name string, created time.Time, condition batchv1.JobConditionType) batchv1.Job {
	job := batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: name, CreationTimestamp: metav1.NewTime(created)}}
	if condition != "" {
		job.Status.Conditions = []batchv1.JobCondition{{Type: condition, Status: corev1.ConditionTrue}}
	}
	return job
}

func TestLatestFinishedJob(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		jobs []batchv1.Job
		want string
	}{
		{name: "empty", want: ""},
		{
			name: "only running",
			jobs: []batchv1.Job{newJob("running", now, "")},
			want: "",
		},
		{
			name: "failed then complete",
			jobs: []batchv1.Job{
				newJob("complete", now, batchv1.JobComplete),
				newJob("failed", now.Add(-time.Hour), batchv1.JobFailed),
			},
			want: "complete",
		},
		{
			name: "complete then failed, ignoring running",
			jobs: []batchv1.Job{
				newJob("complete", now.Add(-time.Hour), batchv1.JobComplete),
				newJob("running", now, ""),
				newJob("failed", now.Add(-time.Minute), batchv1.JobFailed),
			},
			want: "failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LatestFinishedJob(&batchv1.JobList{Items: tt.jobs})
			gotName := ""
			if got != nil {
				gotName = got.Name
			}
			if gotName != tt.want {
				t.Errorf("expected %q, got %q", tt.want, gotName)
			}
		})
	}
}
