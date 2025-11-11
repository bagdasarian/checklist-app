package postgres

import (
    "time"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func convertToTimestamp(t interface{}) *timestamppb.Timestamp {
    switch v := t.(type) {
    case time.Time:
        return timestamppb.New(v)
    case *time.Time:
        if v != nil {
            return timestamppb.New(*v)
        }
    }
    return nil
}