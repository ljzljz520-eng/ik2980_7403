package codec

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"emergencycomms/internal/model"
)

func Fragment(message model.ProtectedMessage, width int) ([]model.Segment, error) {
	if err := message.Validate(); err != nil {
		return nil, err
	}
	if width < 1 {
		return nil, fmt.Errorf("fragment width must be positive")
	}
	parts := make([]string, 0)
	for start := 0; start < len(message.Body); start += width {
		end := start + width
		if end > len(message.Body) {
			end = len(message.Body)
		}
		parts = append(parts, message.Body[start:end])
	}
	if len(parts) == 0 {
		parts = append(parts, "")
	}
	id := message.ChannelNumber + "/" + message.BatchNumber + "/" + message.Nonce
	segments := make([]model.Segment, 0, len(parts))
	for index, payload := range parts {
		segments = append(segments, model.Segment{MessageID: id, Index: index, Total: len(parts), Payload: payload, Checksum: checksum(payload)})
	}
	return segments, nil
}

func Assemble(segments []model.Segment) (string, error) {
	if len(segments) == 0 {
		return "", fmt.Errorf("no segments supplied")
	}
	ordered := append([]model.Segment(nil), segments...)
	total := ordered[0].Total
	messageID := ordered[0].MessageID
	if total != len(ordered) || total < 1 {
		return "", fmt.Errorf("segment total is incomplete")
	}
	for i := range ordered {
		if ordered[i].MessageID != messageID || ordered[i].Total != total || ordered[i].Index != i {
			return "", fmt.Errorf("segment order is invalid")
		}
		if checksum(ordered[i].Payload) != ordered[i].Checksum {
			return "", fmt.Errorf("segment checksum mismatch")
		}
	}
	parts := make([]string, 0, len(ordered))
	for _, segment := range ordered {
		parts = append(parts, segment.Payload)
	}
	return strings.Join(parts, ""), nil
}

func checksum(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:6])
}

func FrameLabel(segment model.Segment) string {
	return fmt.Sprintf("%s:%d/%d", segment.MessageID, segment.Index+1, segment.Total)
}

func EncodeFrame(segment model.Segment) string {
	return strings.Join([]string{segment.MessageID, strconv.Itoa(segment.Index), strconv.Itoa(segment.Total), segment.Payload, segment.Checksum}, "#")
}

func DecodeFrame(value string) (model.Segment, error) {
	fields := strings.Split(value, "#")
	if len(fields) != 5 {
		return model.Segment{}, fmt.Errorf("frame requires five fields")
	}
	index, err := strconv.Atoi(fields[1])
	if err != nil || index < 0 {
		return model.Segment{}, fmt.Errorf("invalid frame index")
	}
	total, err := strconv.Atoi(fields[2])
	if err != nil || total < 1 {
		return model.Segment{}, fmt.Errorf("invalid frame total")
	}
	return model.Segment{MessageID: fields[0], Index: index, Total: total, Payload: fields[3], Checksum: fields[4]}, nil
}

func ValidateFrames(segments []model.Segment) error {
	if len(segments) == 0 {
		return fmt.Errorf("frame list is empty")
	}
	for _, segment := range segments {
		if segment.MessageID == "" || segment.Total < 1 || segment.Index < 0 || segment.Index >= segment.Total {
			return fmt.Errorf("frame metadata invalid")
		}
	}
	return nil
}
