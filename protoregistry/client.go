package protoregistry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/nakji-network/connector/kafkautils"
)

var errProtoFound = errors.New("proto file found")

// TopicTypes is a map where topic schemas are keys and proto.Message are values.
type TopicTypes map[string]proto.Message

// RegisterDynamicTopics registers kafka topic and protobuf type mappings
func RegisterDynamicTopics(host string, topicTypes map[string]proto.Message, msgType kafkautils.MsgType) error {
	tpmList := buildTopicProtoMsgs(topicTypes, msgType)

	err := generateDescriptorFiles(tpmList)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate descriptor files")
	}

	u := url.URL{Scheme: "http", Host: host, Path: "/v1/register"}

	b, err := json.Marshal(tpmList)
	if err != nil {
		return err
	}

	res, err := http.Post(u.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	statusOK := res.StatusCode >= 200 && res.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("call failed, err: %v", string(b))
	}

	return nil
}

type TopicProtoMsg struct {
	MsgType      kafkautils.MsgType `json:"msg_type" binding:"required"`
	TopicName    string             `json:"topic" binding:"required"`
	ProtoMsgName string             `json:"proto_msg" binding:"required"`
	Descriptor   []byte             `json:"descriptor" binding:"required"`
}

type TopicSubscription struct {
	TopicProtoMsgs []TopicProtoMsg
	ShouldUpdate   bool `json:"should_update"`
}

func buildTopicProtoMsgs(topicTypes TopicTypes, msgType kafkautils.MsgType) []*TopicProtoMsg {
	var tpmList []*TopicProtoMsg

	for k, v := range topicTypes {
		t := reflect.TypeOf(v)
		pmn := t.String()

		if t.Kind() == reflect.Ptr {
			elem := t.Elem()
			pmn = elem.String()
		}

		tpm := &TopicProtoMsg{
			MsgType:      msgType,
			TopicName:    k,
			ProtoMsgName: pmn,
		}

		tpmList = append(tpmList, tpm)
	}
	return tpmList
}

// generateDescriptorFiles scans the local disk looking for proto files and generates proto descriptor files.
func generateDescriptorFiles(tpmList []*TopicProtoMsg) error {
	wd, err := os.Getwd()
	if err != nil {
		log.Error().Err(err).Msg("failed to call os.Getwd()")
		return err
	}

	for _, tpm := range tpmList {
		path, err := getProtoFilePath(wd, tpm)
		if err != nil {
			log.Error().Err(err).Interface("tpm", tpm).Msg("failed to get the proto file path")
			return err
		}

		descFile, err := generateDescriptorFile(path)
		if err != nil {
			log.Error().Str("path", path).Str("descFile", descFile).Err(err).Msg("failed to generate descriptor file")
			return err
		}

		desc, err := ioutil.ReadFile(descFile)
		if err != nil {
			log.Error().Err(err).Msg("failed to read descriptor file")
			return err
		}

		tpm.Descriptor = desc
	}

	return nil
}

func generateDescriptorFile(path string) (string, error) {
	descFile := path + ".desc"
	file := filepath.Base(path)
	dir := filepath.Dir(path)

	_, err := os.Stat(descFile)
	if err == nil {
		log.Debug().Str("descFile", descFile).Msg("proto descriptor already exists, skip")
		return descFile, nil
	}

	cmd := exec.Command(
		"protoc",
		"--include_imports",
		"--descriptor_set_out="+descFile,
		"-I="+dir,
		file,
	)

	log.Debug().Str("path", path).Msg("generateDescriptorFile()")
	log.Debug().Str("descFile", descFile).Msg("generateDescriptorFile()")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return descFile, nil
}

func getProtoFilePath(baseDir string, tpm *TopicProtoMsg) (string, error) {
	pSeg := strings.Split(tpm.ProtoMsgName, kafkautils.TopicContextSeparator)
	pkg := pSeg[0]

	p := ""

	err := filepath.WalkDir(baseDir, func(path string, info os.DirEntry, err error) error {
		// Skip permission denied error
		if err != nil && !strings.Contains(err.Error(), fs.ErrPermission.Error()) {
			log.Error().Err(err).Msg("WalkDirFunc error")
			return err
		}

		// terminate the lookup if the file is found
		if !info.IsDir() && filepath.Base(path) == pkg+".proto" {
			p = path
			log.Debug().Str("path", path).Msg("found the proto file")
			return errProtoFound
		}

		return nil
	})

	if err == errProtoFound {
		return p, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("WalkDir error")
		return "", err
	}

	return "", errors.New("proto file not found")
}
