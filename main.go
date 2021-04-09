package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"test-paddle/data"
	"test-paddle/paddle"
)

func GetNewPredictor() *paddle.Predictor {
	config := paddle.NewAnalysisConfig()
	config.SetModel("model/paddle_esmm_model_v1", "")
	// 输出模型路径
	config.DisableGlogInfo()
	config.SwitchUseFeedFetchOps(false)
	config.SwitchSpecifyInputNames(true)
	config.SwitchIrOptim(false)

	return paddle.NewPredictor(config)
}

func main() {

	ch := make(chan *paddle.Predictor, 30)

	for i := 0; i < 30; i++ {
		ch <- GetNewPredictor()
	}

	features := make(map[string][]int64)
	json.Unmarshal([]byte(data.TestData), &features)


	for i := 0; i < 10000; i++ {
		fmt.Printf("i = %+v \n", i)
		predict := <-ch
		go func(ch chan *paddle.Predictor, p *paddle.Predictor) {

			defer func() {
				ch <- p
			}()
			inputs := p.GetInputTensors()
			for _, input := range inputs {
				input.SetValue(features[input.Name()])
				input.Reshape([]int32{int32(len(features[input.Name()])), 1})
				p.SetZeroCopyInput(input)
			}
			p.ZeroCopyRun()
			outputs := p.GetOutputTensors()

			output := outputs[2]
			p.GetZeroCopyOutput(output)

			outputVal := output.Value()
			value := reflect.ValueOf(outputVal)
			tmp := value.Interface().([][]float32)

			var result []float64
			for _, v := range tmp {
				result = append(result, float64(v[0]))
			}
			//fmt.Printf("result = %+v", result)
		}(ch, predict)

	}

}
