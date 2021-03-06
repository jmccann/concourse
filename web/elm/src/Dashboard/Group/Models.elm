module Dashboard.Group.Models exposing (Group, Pipeline)

import Concourse
import Concourse.PipelineStatus as PipelineStatus
import Dashboard.Group.Tag as Tag


type alias Group =
    { pipelines : List Pipeline
    , teamName : String
    , tag : Maybe Tag.Tag
    }


type alias Pipeline =
    { id : Int
    , name : String
    , teamName : String
    , public : Bool
    , jobs : List Concourse.Job
    , resourceError : Bool
    , status : PipelineStatus.PipelineStatus
    , isToggleLoading : Bool
    , isVisibilityLoading : Bool
    }
