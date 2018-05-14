package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type MosResponse struct {
	Time          string  `json:"time"`
	OriginalFile  string  `json:"original file"`
	DegradedFile  string  `json:"degraded file"`
	LinesPerFrame uint16  `json:"Lines per Frame"`
	PixPerLine    uint16  `json:"Pix. Per Line"`
	RefEstFPS     float32 `json:"Ref. Estimated FPS"`
	TestEstFPS    float32 `json:"Test Estimated FPS [Hz]"`
	FrameFreeze   float32 `json:"Frame Freeze [%]"`
	FrameSkip     float32 `json:"Frame Skip [%]"`
	Blur          float32 `json:"Blur [0..10]"`
	Blockiness    float32 `json:"Blockiness [0..10]"`
	Mos           float32 `json:"PEVQ MOSPEVQ"`
}
type TwilioResponseRecordings struct {
	Recordings []Recording_info `json:"recordings"`
	Meta       Meta             `json:"meta"`
}
type Recording_info struct {
	Status          string       `json:"status"`
	GroupingSids    GroupingSids `json:"grouping_sids"`
	ContainerFormat string       `json:"container_format"`
	TrackName       string       `json:"tarck_name"`
	AccountSid      string       `json:"account_sid"`
	Url             string       `json:"url"`
	Codec           string       `json:"codec"`
	SourceSid       string       `json:"source_sid"`
	Sid             string       `json:"sid"`
	Duration        uint32       `json:"duration"`
	DateCreated     string       `json:"date_created"`
	Type            string       `json:"type"`
	Size            uint64       `json:"size"`
	Links           Links        `json:"links"`
}
type GroupingSids struct {
	ParticipantSid string `json:"participant_sid"`
	RoomSid        string `json:"room_sid"`
}
type Links struct {
	Media string `json:"media"`
}
type Meta struct {
	Page            uint32 `json:"page"`
	PageSize        uint32 `json:"page_size"`
	FirstPageUrl    string `json:"first_page_url"`
	PreviousPageUrl string `json:"previous_page_url"`
	Url             string `json:"url"`
	NextPageUrl     string `json:"next_page_url"`
	Key             string `json:"key"`
}

func main() {

	//docker stuff

	cmdNight := "nightwatch"
	argsNight := []string{"script.js", "--config", "conf.json", "--verbose", "-e", "chrome"}
	if err := exec.Command(cmdNight, argsNight...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Nightwatch script ended")

	// Twilio requests + file moving in the project folder
	req, err := http.NewRequest("GET", "https://video.twilio.com/v1/Recordings/", nil)
	if err != nil {
		err = fmt.Errorf("bad request, error")
	}
	req.Header.Set("Authorization", "Basic QUM2MDQ5ZTc1YmVkZjhkZTY1NTg4YjdkMDM4NDBmNjZiMjo4MDU4MTZhNDdiYTk1NTFmNDI5ODJmOGVkODM0MmEzMA==")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Postman-Token", "bbd4da61-5018-4696-6d96-e3834ec24e9c")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("bad response, error")
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", resp.Status)
	}

	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if nil != err {
			fmt.Println("errorination happened reading the body", err)
		}

		s := string(body[:])
		data := &TwilioResponseRecordings{}
		err = json.Unmarshal([]byte(s), data)
		fmt.Println(err)

		RoomSidGet := data.Recordings[0].GroupingSids.RoomSid
		fmt.Println(RoomSidGet)

		RoomLink := "https://video.twilio.com/v1/Recordings/?GroupingSid=" + RoomSidGet
		fmt.Println(RoomLink)
		RoomCloseLink := "https://video.twilio.com/v1/Rooms/" + RoomSidGet
		fmt.Println(RoomCloseLink)

		req1, err := http.NewRequest("GET", RoomLink, nil)
		if err != nil {
			err = fmt.Errorf("bad request, error")
		}
		req1.Header.Set("Authorization", "Basic QUM2MDQ5ZTc1YmVkZjhkZTY1NTg4YjdkMDM4NDBmNjZiMjo4MDU4MTZhNDdiYTk1NTFmNDI5ODJmOGVkODM0MmEzMA==")
		req1.Header.Set("Cache-Control", "no-cache")
		req1.Header.Set("Postman-Token", "bbd4da61-5018-4696-6d96-e3834ec24e9c")

		resp1, err := http.DefaultClient.Do(req1)
		if err != nil {
			err = fmt.Errorf("bad response, error")
		}

		if resp1.StatusCode != http.StatusOK {
			err = fmt.Errorf("bad status: %s", resp1.Status)
		}

		if resp1.StatusCode == http.StatusOK {
			defer resp1.Body.Close()
			body1, err := ioutil.ReadAll(resp1.Body)

			if nil != err {
				fmt.Println("errorination happened reading the body", err)
			}
			s1 := string(body1[:])
			data1 := &TwilioResponseRecordings{}
			err = json.Unmarshal([]byte(s1), data1)
			fmt.Println(err)

			mediaLink := string(' ')

			if data1.Recordings[0].Type == "video" {
				mediaLink = data1.Recordings[0].Links.Media
			}
			if data1.Recordings[1].Type == "video" {
				mediaLink = data1.Recordings[1].Links.Media
			}

			// Close Twilio Room

			payload := strings.NewReader("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"Status\"\r\n\r\ncompleted\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--")

			reqCompl, err := http.NewRequest("POST", RoomCloseLink, payload)
			if err != nil {
				err = fmt.Errorf("bad request to Twilio Close Room")
			}
			reqCompl.Header.Add("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")
			reqCompl.Header.Add("Authorization", "Basic QUM2MDQ5ZTc1YmVkZjhkZTY1NTg4YjdkMDM4NDBmNjZiMjo4MDU4MTZhNDdiYTk1NTFmNDI5ODJmOGVkODM0MmEzMA==")
			reqCompl.Header.Add("Cache-Control", "no-cache")
			reqCompl.Header.Add("Postman-Token", "70d0ada7-5c88-8127-127f-47c2dcdfbd4b")

			respCompl, err := http.DefaultClient.Do(reqCompl)
			if err != nil {
				err = fmt.Errorf("bad response, error")
			}

			if respCompl.StatusCode != http.StatusOK {
				err = fmt.Errorf("bad status: %s", respCompl.Status)
			}
			if respCompl.StatusCode == http.StatusOK {
				defer respCompl.Body.Close()
				bodyCompl, err := ioutil.ReadAll(respCompl.Body)
				if nil != err {
					fmt.Println("errorination happened reading the body", err)
				}
				sCompl := string(bodyCompl[:])
				fmt.Println(sCompl)

			}

			//WGET MEDIA from Twilio

			fmt.Println(mediaLink)
			// Waiting for Twilio file upload then request the recording
			time.Sleep(20000 * time.Millisecond)

			cmdWget := "wget"
			argsWget := []string{
				"--http-user=AC6049e75bedf8de65588b7d03840f66b2",
				"--http-password=805816a47ba9551f42982f8ed8342a30",
				mediaLink}

			if err := exec.Command(cmdWget, argsWget...).Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Println("Video file downloaded from twilio as Media")
			time.Sleep(5000 * time.Millisecond)

			cmdExt := "mv"
			argsExt := []string{"Media", "_twilioVideo.webm"}

			if err := exec.Command(cmdExt, argsExt...).Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Println("Video file extensiton changed new file : twilioVideo.webm")

			cmdMvRemote := "mv"
			argsMvRemote := []string{"../../../Downloads/remoteRec.webm", "_remoteRec.webm"}

			if err := exec.Command(cmdMvRemote, argsMvRemote...).Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Println("Video file remoteRec.webm has been moved to project folder")

			cmdTwilioFPS := exec.Command("ffmpeg", "-y", "-i", "_twilioVideo.webm", "-vf", "setpts=1.25*PTS", "-r", "24", "_twilioVideo24.webm")

			out, err := cmdTwilioFPS.CombinedOutput()
			if err != nil {
				fmt.Printf("cmdTwilioFPS.Run() failed with %s\n", err.Error())
				fmt.Println(string(out))
			}
			fmt.Println("FFmpeg Change FPS for twilio recording")

			// MOS SERVER reqest

			fmt.Println("Starting MOS")

			var client *http.Client
			var remoteURL string
			{
				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					b, err := httputil.DumpRequest(r, true)
					if err != nil {
						panic(err)
					}
					fmt.Printf("%s", b)
				}))
				defer ts.Close()
				client = ts.Client()
				remoteURL = "http://10.1.5.58:8080/MOS/UploadServlet2"
			}

			// //prepare the reader instances to encode
			values := map[string]io.Reader{

				"1": mustOpen("_twilioVideo24.webm"), // files
				"2": mustOpen("_remoteRec.webm"),
			}

			err = Upload(client, remoteURL, values)
			if err != nil {
				panic(err)
			}

		}
	}

}

func Upload(client *http.Client, url string, values map[string]io.Reader) (err error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}

	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}

	if res.StatusCode == http.StatusOK {

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)

		if nil != err {
			fmt.Println("errorination happened reading the body", err)
		}

		var bodyContent []byte
		// fmt.Println(res.StatusCode)
		// fmt.Println(res.Header)
		res.Body.Read(bodyContent)
		res.Body.Close()
		// fmt.Println(bodyContent)

		// fmt.Println(string(body[:]))

		s := string(body[:])
		s = strings.Replace(s, "\n", "", -1)
		s = strings.Replace(s, "\r", "", -1)
		data := &MosResponse{}
		err = json.Unmarshal([]byte(s), data)
		fmt.Println(err)
		fmt.Println(data.Mos)
		s2, _ := json.Marshal(data)
		fmt.Println(string(s2))

	}
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}
