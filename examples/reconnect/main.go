package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/XdpCs/comfyUIclient"
)

var workflow = `{
	"3": {
	  "inputs": {
		"seed": 4702551705662,
		"steps": 20,
		"cfg": 8,
		"sampler_name": "euler",
		"scheduler": "normal",
		"denoise": 1,
		"model": [
		  "4",
		  0
		],
		"positive": [
		  "6",
		  0
		],
		"negative": [
		  "7",
		  0
		],
		"latent_image": [
		  "5",
		  0
		]
	  },
	  "class_type": "KSampler",
	  "_meta": {
		"title": "KSampler"
	  }
	},
	"4": {
	  "inputs": {
		"ckpt_name": "v1-5-pruned-emaonly.safetensors"
	  },
	  "class_type": "CheckpointLoaderSimple",
	  "_meta": {
		"title": "Load Checkpoint"
	  }
	},
	"5": {
	  "inputs": {
		"width": 512,
		"height": 512,
		"batch_size": 1
	  },
	  "class_type": "EmptyLatentImage",
	  "_meta": {
		"title": "Empty Latent Image"
	  }
	},
	"6": {
	  "inputs": {
		"text": "beautiful scenery nature glass bottle landscape, , purple galaxy bottle,",
		"clip": [
		  "4",
		  1
		]
	  },
	  "class_type": "CLIPTextEncode",
	  "_meta": {
		"title": "CLIP Text Encode (Prompt)"
	  }
	},
	"7": {
	  "inputs": {
		"text": "text, watermark",
		"clip": [
		  "4",
		  1
		]
	  },
	  "class_type": "CLIPTextEncode",
	  "_meta": {
		"title": "CLIP Text Encode (Prompt)"
	  }
	},
	"8": {
	  "inputs": {
		"samples": [
		  "3",
		  0
		],
		"vae": [
		  "4",
		  2
		]
	  },
	  "class_type": "VAEDecode",
	  "_meta": {
		"title": "VAE Decode"
	  }
	},
	"9": {
	  "inputs": {
		"filename_prefix": "ComfyUI",
		"images": [
		  "8",
		  0
		]
	  },
	  "class_type": "SaveImage",
	  "_meta": {
		"title": "Save Image"
	  }
	}
  }`

var baseURL = "http://127.0.0.1:6889"

func main() {
	cli, err := comfyUIclient.NewReConnectClient(baseURL, &http.Client{Timeout: time.Second})
	if err != nil {
		return
	}
	cli.Start()

	for !cli.IsInitialized() {
		time.Sleep(time.Second)
	}

	workflow = comfyUIclient.TraverseAndModifySeed(workflow)

	resp, err := cli.QueuePromptByString(workflow, "")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)

	count := 1
	for taskStatus := range cli.Watch() {
		fmt.Println(taskStatus.Type)
		switch taskStatus.Type {
		case comfyUIclient.ExecutionStart:
			s := taskStatus.Data.(*comfyUIclient.WSMessageDataExecutionStart)
			fmt.Printf("Type: %v, Data:%+v\n", comfyUIclient.ExecutionStart, s)
		case comfyUIclient.ExecutionCached:
			s := taskStatus.Data.(*comfyUIclient.WSMessageDataExecutionCached)
			fmt.Printf("Type: %v, Data:%+v\n", comfyUIclient.ExecutionCached, s)
		case comfyUIclient.Executing:
			s := taskStatus.Data.(*comfyUIclient.WSMessageDataExecuting)
			fmt.Printf("Type: %v, Data:%+v\n", comfyUIclient.Executing, s)
		case comfyUIclient.Progress:
			s := taskStatus.Data.(*comfyUIclient.WSMessageDataProgress)
			fmt.Printf("Type: %v, Data:%+v\n", comfyUIclient.Progress, s)
		case comfyUIclient.Executed:
			s := taskStatus.Data.(*comfyUIclient.WSMessageDataExecuted)
			fmt.Printf("Type: %v, Data:%+v\n", comfyUIclient.Executed, s)
			for _, images := range s.Output {
				for _, image := range images {
					imageData, err := cli.GetFile(image)
					if err != nil {
						panic(err)
					}
					f, err := os.Create(image.Filename)
					if err != nil {
						log.Println("Failed to write image:", err)
						os.Exit(1)
					}
					f.Write(*imageData)
					f.Close()
				}
			}
			count++
			fmt.Println("============================")
			workflow = comfyUIclient.TraverseAndModifySeed(workflow)
			cli.QueuePromptByString(workflow, "")
			if count == 3 {
				cli.Close()
			}
		case comfyUIclient.ExecutionInterrupted:
			s := taskStatus.Data.(*comfyUIclient.WSMessageExecutionInterrupted)
			fmt.Printf("Type: %v, Data:%+v\n", comfyUIclient.ExecutionInterrupted, s)
			count++
			IsEndQueuePrompt(count, 2)
		case comfyUIclient.ExecutionError:
			s := taskStatus.Data.(*comfyUIclient.WSMessageExecutionError)
			fmt.Printf("Type: %v, Data:%+v\n", comfyUIclient.ExecutionError, s)
			count++
			IsEndQueuePrompt(count, 2)
		default:
			fmt.Println("unknown message type")
		}
	}
}

func IsEndQueuePrompt(count int, num int) {

}
