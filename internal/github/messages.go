package github

import (
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	viewPush     = "View Push"
	viewPR       = "View #%#d"
	viewReview   = "View Review"
	viewOnGithub = "View on GitHub"
	themeColor   = "#6e5494"

	msgRepeatCommitMessageLink = "%#s [\U0001f50D](%s)"
	msgNewCommitMessageLink    = "%#+s [\U0001f50D](%s)"
	msgCompareBaseToBranch     = "Compare %s...%s"
	msgUserCreatedTag          = "%#+s created tag %#+s"
	msgUserDeletedBranch       = "%#+s deleted branch %#+s"
	msgUserDeletedTag          = "%#+s deleted tag %#+s"
	msgUserVerbedPRBranch      = "%#+s %m #%#d: %#+s into %#+s"
	msgUserVerbedPRTitle       = "%#+s %m #%#d: %#+s"
	msgUserReviewState         = "%#s %m [#%#d: %#s](%s)"
	msgUserEditedReview        = "%#+s edited a review on **#%#d**"
	msgUserDismissedReview     = "%#+s dismissed a review on **#%#d**"
	msgUserSubmittedReview     = "%#+s submitted a review on **#%#d**"
	msgUserCommentedOn         = "%#+s commented on **#%#d**"
	msgVerbedPR                = "verbed pr"
	msgReviewedPR              = "reviewed pr"
	msgCommentedPR             = "commented pr"
	msgEditedReview            = "edited review"
	msgDismissedReview         = "dismissed review"
	msgWorkflowStatusSummary   = "status||job|summary"
	msgWorkflowStatus          = "status||job"
	msgWorkflowDetail          = "detail||job"

	prChangesRequested    = "changes_requested||review"
	prEditedReview        = "edited||review"
	prDismissedReview     = "dismissed||review"
	prCommented           = "commented||review"
	prOpened              = "opened||pr"
	prOpenedDraft         = "opened||pr|draft"
	prClosed              = "closed||pr|"
	prClosedMerged        = "closed||pr|merged"
	prOpenedSummary       = "opened||pr|summary"
	prOpenedDraftSummary  = "opened||pr|draft|summary"
	prClosedSummary       = "closed||pr|summary"
	prClosedMergedSummary = "closed||pr|merged|summary"
	createTag             = "tag||create"
	deleteTag             = "tag||delete"
	deleteBranch          = "branch||delete"

	branchPushed      = "pushed"
	branchForced      = "forced"
	branchPushText    = "pushed||branch"
	branchPushSummary = "pushed||branch|summary"

	jobSuccess         = "success||job"
	jobFailure         = "failure||job"
	jobCancelled       = "cancelled||job"
	jobSkipped         = "skipped||job"
	jobSuccessSymbol   = "success||job|sym"
	jobFailureSymbol   = "failure||job|sym"
	jobCancelledSymbol = "cancelled||job|sym"
	jobSkippedSymbol   = "skipped||job|sym"
)

func init() {
	_ = message.SetString(language.English, branchPushSummary, "%s %m %s")
	_ = message.Set(language.English, branchPushText, plural.Selectf(3, "%d",
		plural.One, "%#+s %m %d commit to %#+s",
		plural.Other, "%#+s %m %d commits to %#+s"))
	_ = message.SetString(language.English, branchPushed, "pushed")
	_ = message.SetString(language.English, branchForced, "force-pushed")
	_ = message.SetString(language.English, prChangesRequested, "requested changes for")
	_ = message.SetString(language.English, prEditedReview, "edited a review of")
	_ = message.SetString(language.English, prDismissedReview, "dismissed a review of")
	_ = message.SetString(language.English, prCommented, "commented on")
	_ = message.SetString(language.English, prOpened, "opened pull request")
	_ = message.SetString(language.English, prOpenedDraft, "opened draft pull request")
	_ = message.SetString(language.English, prClosed, "closed pull request")
	_ = message.SetString(language.English, prClosedMerged, "merged pull request")
	_ = message.SetString(language.English, prOpenedSummary, "opened PR")
	_ = message.SetString(language.English, prOpenedDraftSummary, "opened draft PR")
	_ = message.SetString(language.English, prClosedSummary, "closed PR")
	_ = message.SetString(language.English, prClosedMergedSummary, "merged PR")
	_ = message.SetString(language.English, createTag, "%s tagged %s")
	_ = message.SetString(language.English, deleteBranch, "%s deleted %s")
	_ = message.SetString(language.English, deleteTag, "%s untagged %s")
	_ = message.SetString(language.English, msgVerbedPR, "%s %m #%#d")
	_ = message.SetString(language.English, msgReviewedPR, "%s reviewed #%#d")
	_ = message.SetString(language.English, msgCommentedPR, "%s commented on #%#d")
	_ = message.SetString(language.English, msgEditedReview, "%s edited #%#d review")
	_ = message.SetString(language.English, jobSuccess, "passed")
	_ = message.SetString(language.English, jobFailure, "failed")
	_ = message.SetString(language.English, jobCancelled, "was cancelled")
	_ = message.SetString(language.English, jobSkipped, "was skipped")
	_ = message.SetString(language.English, jobSuccessSymbol, "✔")
	_ = message.SetString(language.English, jobFailureSymbol, "❌")
	_ = message.SetString(language.English, jobCancelledSymbol, "🚫")
	_ = message.SetString(language.English, jobSkippedSymbol, "◌")
}
