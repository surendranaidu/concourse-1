module Dashboard.Group.Models exposing
    ( Group
    , Pipeline
    , PipelineCardStatus(..)
    , isRunning
    , show
    )

import Concourse.PipelineStatus as PipelineStatus exposing (PipelineStatus(..))


type alias Group =
    { pipelines : List Pipeline
    , teamName : String
    }


type alias Pipeline =
    { id : Int
    , name : String
    , teamName : String
    , public : Bool
    , isToggleLoading : Bool
    , isVisibilityLoading : Bool
    , paused : Bool
    , archived : Bool
    }


type PipelineCardStatus
    = PipelineStatusPaused
    | PipelineStatusAborted PipelineStatus.StatusDetails
    | PipelineStatusErrored PipelineStatus.StatusDetails
    | PipelineStatusFailed PipelineStatus.StatusDetails
    | PipelineStatusPending Bool
    | PipelineStatusSucceeded PipelineStatus.StatusDetails
    | PipelineStatusUnknown
    | PipelineStatusJobsDisabled


show : PipelineCardStatus -> String
show status =
    case status of
        PipelineStatusPaused ->
            "paused"

        PipelineStatusAborted _ ->
            "aborted"

        PipelineStatusErrored _ ->
            "errored"

        PipelineStatusFailed _ ->
            "failed"

        PipelineStatusPending _ ->
            "pending"

        PipelineStatusSucceeded _ ->
            "succeeded"

        PipelineStatusUnknown ->
            "unknown"

        PipelineStatusJobsDisabled ->
            ""


isRunning : PipelineCardStatus -> Bool
isRunning status =
    case status of
        PipelineStatusPaused ->
            False

        PipelineStatusAborted details ->
            details == PipelineStatus.Running

        PipelineStatusErrored details ->
            details == PipelineStatus.Running

        PipelineStatusFailed details ->
            details == PipelineStatus.Running

        PipelineStatusPending bool ->
            bool

        PipelineStatusSucceeded details ->
            details == PipelineStatus.Running

        PipelineStatusUnknown ->
            False

        PipelineStatusJobsDisabled ->
            False
