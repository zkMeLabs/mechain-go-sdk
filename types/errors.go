package types

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const unknownErr = "unknown error"

var (
	ErrorDefaultAccountNotExist = errors.New("Default account of client is not exist ")
	ErrorProposalIDNotFound     = errors.New("Proposal ID not found ")
)

// ErrResponse define the information of the error response
type ErrResponse struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`
	Message    string   `xml:"Message"`
	StatusCode int
}

// Error returns the error msg
func (r ErrResponse) Error() string {
	return fmt.Sprintf("statusCode %v : code : %s  (Message: %s)",
		r.StatusCode, r.Code, r.Message)
}

// ConstructErrResponse  checks the response is an error response
func ConstructErrResponse(r *http.Response, bucketName, objectName string) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	if r == nil {
		return ErrResponse{
			StatusCode: r.StatusCode,
			Code:       unknownErr,
			Message:    "Response is empty ",
		}
	}

	errResp := ErrResponse{}
	errResp.StatusCode = r.StatusCode

	// read err body of max 10M size
	const maxBodySize = 10 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		return ErrResponse{
			StatusCode: r.StatusCode,
			Code:       "InternalError",
			Message:    err.Error(),
		}
	}
	// decode the xml content from response body
	decodeErr := xml.NewDecoder(bytes.NewReader(body)).Decode(&errResp)
	if decodeErr != nil {
		switch r.StatusCode {
		case http.StatusNotFound:
			if bucketName != "" {
				if objectName == "" {
					errResp = ErrResponse{
						StatusCode: r.StatusCode,
						Code:       "NoSuchBucket",
						Message:    "The specified bucket does not exist.",
					}
				} else {
					errResp = ErrResponse{
						StatusCode: r.StatusCode,
						Code:       "NoSuchObject",
						Message:    "The specified object does not exist.",
					}
				}
			}
		case http.StatusForbidden:
			errResp = ErrResponse{
				StatusCode: r.StatusCode,
				Code:       "AccessDenied",
				Message:    "no permission to access the resource",
			}
		default:
			errBody := bytes.TrimSpace(body)
			msg := unknownErr
			if len(errBody) > 0 {
				msg = string(errBody)
			}
			fmt.Println("default error msg :", msg)
			errResp = ErrResponse{
				StatusCode: r.StatusCode,
				Code:       unknownErr,
				Message:    msg,
			}
		}
	}

	return errResp
}

// ToInvalidArgumentResp returns invalid argument response.
func ToInvalidArgumentResp(message string) error {
	return ErrResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "InvalidArgument",
		Message:    message,
	}
}
