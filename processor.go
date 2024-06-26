package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

func initWebsocketClient(conf *Config) error {
	if conf == nil {
		return errors.New("config is nil")
	}

	Rlog.SetDebugEnabled(conf.EnableDebugMessages)

	Rlog.Info("Starting Client")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: conf.CLIAddress, Path: fmt.Sprintf("/v1/receive/%s", conf.SelfNumber)}
	Rlog.Infof("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		Rlog.Fatal("dial:", err)
		return err
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			Rlog.Fatal("close:", err)
		}
	}(c)

	Rlog.Infof("ws connected %s", u.String())

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				Rlog.Error("read error: ", err)
				continue
			}

			var msg SignalMessage

			err = json.Unmarshal(message, &msg)
			if err != nil {
				Rlog.Error("decode error: ", err.Error())
				continue
			}

			now := uint64(time.Now().UTC().UnixMilli())
			if (now - msg.Envelope.Timestamp) > conf.IgnoreOlderMessages {
				Rlog.Debugf("Now is %d, but message is from %d; diff is %d (>%d)", now, msg.Envelope.Timestamp, now-msg.Envelope.Timestamp, conf.IgnoreOlderMessages)
				continue //this is sync message, will be ignored
			}

			if conf.IsPrintMessages && (len(msg.Envelope.DataMessage.Message) > 0 || len(msg.Envelope.DataMessage.Attachments) > 0) {
				Rlog.Infof("Message: %s, Author: %s, Author UUID: %s, Attachments: %d, Group: %s",
					msg.Envelope.DataMessage.Message,
					msg.Envelope.Source,
					msg.Envelope.SourceUuid,
					len(msg.Envelope.DataMessage.Attachments),
					msg.Envelope.DataMessage.GroupInfo.GroupId,
				)
			}

			rec, err := GetForwardingRecord(conf, msg.Envelope.DataMessage.GroupInfo.GroupId)
			if err != nil {
				Rlog.Error("GetForwardingRecord:", err)
				continue
			}
			if rec == nil {
				Rlog.Debugf("GroupId %s is not found in forwarding list, ignoring", msg.Envelope.DataMessage.GroupInfo.GroupId)
				continue
			}

			Rlog.Debugf("recv: %s", message)

			switch rec.ForwardingMode {
			case FwModeAttachments:
				if len(msg.Envelope.DataMessage.Attachments) > 0 {
					if !conf.IsSendingEnabled {
						Rlog.Debug("sending messages disabled")
						continue
					}

					m, err := CheckFilters(conf, rec, &msg.Envelope, false)
					if err != nil {
						Rlog.Error("check filters error:", err)
						continue
					}

					if len(m) == 0 {
						Rlog.Debugf("filtered message, ignoring...")
						continue
					}

					err = SendMessage(conf, rec.ReceiversGroupIds, msg.Envelope.DataMessage.Attachments, rec.BotSpecialAddonMsg)
					if err != nil {
						Rlog.Error("send message error:", err)
					}

					err = MarkMessageAsRead(conf, msg.Envelope.Source, msg.Envelope.Timestamp) //TODO: this doesn't has any effect (
					if err != nil {
						Rlog.Error("mark message as read error:", err)
					}

					err = SendMessageReaction(conf, rec.ReactionMark, msg.Envelope.Source, msg.Envelope.Source, msg.Envelope.Timestamp)
					if err != nil {
						Rlog.Error("send message reaction error:", err)
					}
				} else {
					Rlog.Debug("message has no attachments")
					continue
				}
				break
			case FwModeMessages:
				if len(msg.Envelope.DataMessage.Message) > 0 && len(msg.Envelope.DataMessage.Attachments) == 0 {
					if !conf.IsSendingEnabled {
						Rlog.Info("sending messages disabled")
						continue
					}

					m, err := CheckFilters(conf, rec, &msg.Envelope, true)
					if err != nil {
						Rlog.Error("check filters error:", err)
						continue
					}

					err = SendMessage(conf, rec.ReceiversGroupIds, make([]SignalAttachments, 0), m)
					if err != nil {
						Rlog.Error("send message error:", err)
					}

					err = MarkMessageAsRead(conf, msg.Envelope.Source, msg.Envelope.Timestamp) //TODO: this doesn't has any effect (
					if err != nil {
						Rlog.Error("mark message as read error:", err)
					}

					err = SendMessageReaction(conf, rec.ReactionMark, msg.Envelope.Source, msg.Envelope.Source, msg.Envelope.Timestamp)
					if err != nil {
						Rlog.Error("send message reaction error:", err)
					}
				}
				break
			case FwModeAll:
				if !conf.IsSendingEnabled {
					Rlog.Debug("sending messages disabled")
					continue
				}

				m, err := CheckFilters(conf, rec, &msg.Envelope, false)
				if err != nil {
					Rlog.Error("check filters error:", err)
					continue
				}

				if len(m) == 0 {
					Rlog.Debugf("filtered message, ignoring...")
					continue
				}

				err = SendMessage(conf, rec.ReceiversGroupIds, msg.Envelope.DataMessage.Attachments, m)
				if err != nil {
					Rlog.Error("send message error:", err)
				}

				err = MarkMessageAsRead(conf, msg.Envelope.Source, msg.Envelope.Timestamp) //TODO: this doesn't has any effect (
				if err != nil {
					Rlog.Error("mark message as read error:", err)
				}

				err = SendMessageReaction(conf, rec.ReactionMark, msg.Envelope.Source, msg.Envelope.Source, msg.Envelope.Timestamp)
				if err != nil {
					Rlog.Error("send message reaction error:", err)
				}
				break
			}

			if len(msg.Envelope.DataMessage.Attachments) > 0 {
				for _, rep := range conf.Forwarding {
					if !rep.IsEnabled {
						Rlog.Debugf("record for group  %s is disabled, ignoring", rep.GroupId)
						continue
					}

					if strings.EqualFold(rep.GroupId, msg.Envelope.DataMessage.GroupInfo.GroupId) {
						Rlog.Debugf("recv: %s", message)
					} else {
						Rlog.Debugf("rep.GroupId %s != msg.Envelope.DataMessage.GroupInfo.GroupId %s", rep.GroupId, msg.Envelope.DataMessage.GroupInfo.GroupId)
					}
				}
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return err
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second * 5):
			}
			return nil
		}
	}
}

func CheckFilters(conf *Config, cg *ConfigGroup, env *SignalEnvelope, isFilterMessage bool) (msg string, err error) {
	if conf == nil {
		return "", errors.New("config is nil")
	}
	if env == nil {
		return "", errors.New("env is nil")
	}

	findFn := func(Source string, Senders []string) bool {
		for _, sn := range Senders {
			if strings.EqualFold(Source, sn) {
				return true
			}
		}

		return false
	}

	if len(cg.SenderNames) > 0 {
		if !findFn(env.SourceName, cg.SenderNames) {
			Rlog.Debugf("Sender name %s is not in Sender Names list config, ignoring message", env.SourceName)
			return "", nil //nothing to do
		}
	}

	if len(cg.SenderUUIDs) > 0 {
		if !findFn(env.SourceUuid, cg.SenderUUIDs) {
			Rlog.Debugf("Sender UUID %s is not in Sender UUIDs list config, ignoring message", env.SourceUuid)
			return "", nil //nothing to do
		}
	}

	if !isFilterMessage {
		return "ok", nil
	}

	findStartWithFn := func(Msg string, Masks []string) bool {
		for _, m := range Masks {
			if strings.HasPrefix(Msg, m) {
				return true
			}
		}

		return false
	}

	if len(cg.StartsWith) > 0 {
		if !findStartWithFn(env.DataMessage.Message, cg.StartsWith) {
			Rlog.Debugf("Message %s is not in starts with list config, ignoring message", env.DataMessage.Message)
			if len(cg.Contains) == 0 { // return here only if we don't have contains config
				return "", nil //nothing to do
			}
		} else {
			return env.DataMessage.Message, nil
		}
	}

	findContainsFn := func(Msg string, Masks []string) bool {
		for _, m := range Masks {
			if strings.Contains(Msg, m) {
				return true
			}
		}

		return false
	}

	if len(cg.Contains) > 0 {
		if !findContainsFn(env.DataMessage.Message, cg.Contains) {
			Rlog.Debugf("Message %s is not contains list config, ignoring message", env.DataMessage.Message)
			return "", nil //nothing to do
		}
	}

	return env.DataMessage.Message, nil
}

func GetForwardingRecord(conf *Config, groupId string) (*ConfigGroup, error) {
	for _, rep := range conf.Forwarding {
		if !rep.IsEnabled {
			Rlog.Debugf("record for group %s is disabled, ignoring", rep.GroupId)
			continue
		}

		if strings.EqualFold(rep.GroupId, groupId) {
			return &rep, nil
		}
	}

	return nil, nil
}

func SendMessage(conf *Config, recGroupIds []string, attachments []SignalAttachments, msgText string) error {
	if conf == nil {
		return errors.New("config is nil")
	}
	if !conf.IsSendingEnabled {
		Rlog.Infof("sending messages disabled")
		return nil
	}
	if len(attachments) == 0 && len(msgText) == 0 {
		return nil
	}

	var msg SignalSendMessageV2

	msg.Message = msgText
	msg.Number = conf.SelfNumber

	for _, rec := range recGroupIds {
		r := rec
		if !strings.HasPrefix(rec, "group.") {
			r = fmt.Sprintf("group.%s", base64.StdEncoding.EncodeToString([]byte(rec)))
		}
		msg.Recipients = append(msg.Recipients, r)
	}

	msg.Mentions = make([]SignalMessageMentions, 0)
	msg.QuoteMentions = make([]SignalMessageMentions, 0)
	msg.Base64Attachments = make([]string, len(attachments))

	for i, attachment := range attachments {
		response, err := http.Get(fmt.Sprintf("http://%s/v1/attachments/%s", conf.CLIAddress, attachment.Id))
		if err != nil {
			Rlog.Error("attachment error: ", err.Error())
			return err
		}
		pr, pw := io.Pipe()
		encoder := base64.NewEncoder(base64.StdEncoding, pw)
		go func() {
			_, err := io.Copy(encoder, response.Body)
			encoder.Close()

			if err != nil {
				pw.CloseWithError(err)
			} else {
				pw.Close()
			}
		}()
		out, err := io.ReadAll(pr)
		if err != nil {
			Rlog.Error("read error: ", err.Error())
			return err
		}

		msg.Base64Attachments[i] = fmt.Sprintf("data:%s;filename=%s;base64,%s", attachment.ContentType, attachment.Filename, string(out))
	}

	resp, err := json.Marshal(msg)
	if err != nil {
		Rlog.Error("json marshal err: ", err)
		return err
	}

	//log.Println("message data: ", string(resp))

	r, err := http.NewRequest("POST", fmt.Sprintf("http://%s/v2/send", conf.CLIAddress), bytes.NewBuffer(resp))
	if err != nil {
		Rlog.Error("new request err: ", err)
		return err
	}

	r.Header.Add("Content-Type", "application/json")
	Rlog.Infof("SENDING MESSAGE TO %s", strings.Join(recGroupIds, ","))
	client := &http.Client{}
	res, err := client.Do(r)
	defer res.Body.Close()
	if err != nil {
		Rlog.Error("client send request error: ", err)
		return err
	}
	response := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		Rlog.Error("client send request resp decode error: ", err)
		return err
	}

	Rlog.Info("Message sent: ", response)

	return nil
}

func MarkMessageAsRead(conf *Config, recipient string, timestamp uint64) error {
	//send receipt
	request := make(map[string]interface{})
	request["receipt_type"] = "read"
	request["recipient"] = recipient
	request["timestamp"] = timestamp

	resp, err := json.Marshal(request)
	if err != nil {
		Rlog.Error("json marshal err: ", err)
		return err
	}
	r, err := http.NewRequest("POST", fmt.Sprintf("http://%s/v1/receipts/%s", conf.CLIAddress, conf.SelfNumber), bytes.NewBuffer(resp))
	if err != nil {
		Rlog.Error("new request err: ", err)
		return err
	}
	r.Header.Add("Content-Type", "application/json")
	Rlog.Infof("MARKING MESSAGE %d AS READ", timestamp)
	client := &http.Client{}
	res, err := client.Do(r)
	defer res.Body.Close()
	if err != nil {
		Rlog.Error("client send request error: ", err)
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return errors.New("receipt bad status")
	}
	return nil
}

func SendMessageReaction(conf *Config, reactionMark string, recipient string, targetAuthor string, timestamp uint64) error {
	if len(reactionMark) == 0 {
		return nil //nothing to do
	}
	//send reaction
	request := make(map[string]interface{})
	request["reaction"] = reactionMark
	request["recipient"] = recipient
	request["target_author"] = targetAuthor
	request["timestamp"] = timestamp

	resp, err := json.Marshal(request)
	if err != nil {
		Rlog.Error("json marshal err: ", err)
		return err
	}
	r, err := http.NewRequest("POST", fmt.Sprintf("http://%s/v1/reactions/%s", conf.CLIAddress, conf.SelfNumber), bytes.NewBuffer(resp))
	if err != nil {
		Rlog.Error("new request err: ", err)
		return err
	}
	r.Header.Add("Content-Type", "application/json")
	Rlog.Infof("MARKING MESSAGE %d WITH REACTION %s", timestamp, reactionMark)
	client := &http.Client{}
	res, err := client.Do(r)
	defer res.Body.Close()
	if err != nil {
		Rlog.Error("client send request error: ", err)
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return errors.New("receipt bad status")
	}
	return nil
}

func GetGroupsList(conf *Config) ([]SignalGroupEntry, error) {
	response, err := http.Get(fmt.Sprintf("http://%s/v1/groups/%s", conf.CLIAddress, conf.SelfNumber))
	if err != nil {
		Rlog.Error("groups list error: ", err.Error())
		return nil, err
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		Rlog.Error("groups io read all error: ", err.Error())
		return nil, err
	}

	if response.StatusCode != 200 {
		resp := make(map[string]string)
		err = json.Unmarshal(body, &resp)
		if err != nil {
			Rlog.Error("err response json unmarshal error: ", err.Error())
			return nil, err
		}
		e := ""
		for k, v := range resp {
			e = fmt.Sprintf("%s: %s", k, v)
		}

		return nil, errors.New(e)
	}

	var groups []SignalGroupEntry
	err = json.Unmarshal(body, &groups)
	if err != nil {
		Rlog.Error("groups json unmarshal error: ", err.Error())
		return nil, err
	}

	return groups, nil
}
