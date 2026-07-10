package tkPresentationMiddlewareHoneypot

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

type AiTrapGenerator struct {
	mutex     sync.Mutex
	randomGen *rand.Rand
}

func NewAiTrapGenerator() *AiTrapGenerator {
	return &AiTrapGenerator{
		randomGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (generator *AiTrapGenerator) Generate(size int) string {
	if size <= 0 {
		return ""
	}

	var result strings.Builder
	result.Grow(size)

	for result.Len() < size {
		pattern := generator.resolveHallucinationPattern()
		remaining := size - result.Len()
		if len(pattern) > remaining {
			pattern = pattern[:remaining]
		}
		result.WriteString(pattern)
		if result.Len() < size {
			result.WriteByte('\n')
		}
	}

	return result.String()
}

func (generator *AiTrapGenerator) resolveHallucinationPattern() string {
	patterns := []string{
		"model_checkpoint_epoch_42_loss_0.0034_accuracy_0.9987_weights_updated",
		"embedding_vector_dim_768_identifier_4521_cosine_similarity_0.8923",
		"training_batch_1847_gradient_norm_0.0012_learning_rate_1e-5_decay_0.99",
		"inference_latency_23ms_sequences_per_second_847_cache_hit_ratio_0.94",
		"fine_tune_dataset_split_train_0.8_val_0.1_test_0.1_samples_124000",
		"attention_head_12_layer_6_activation_relu_dropout_0.1_norm_layer",
		"loss_function_cross_entropy_optimizer_adam_weight_decay_0.01_momentum_0.9",
		"vocabulary_size_50257_max_sequence_length_2048_padding_identifier_0",
		"evaluation_bleu_score_34.7_rouge_l_0.42_perplexity_12.3_humaneval_0.67",
		"checkpoint_path_models/v2/ckpt-1847-shard-003-of-016.safetensors",
		"hyperparameter_search_space_lr_1e-6_to_1e-3_batch_8_to_64_layers_6_to_24",
		"data_pipeline_preprocessing_segmentation_deduplication_shuffling_seed_42",
		"quantization_int8_kv_cache_fp16_flash_attention_enabled_batch_size_32",
		"deployment_config_max_concurrent_requests_128_timeout_30s_retry_backoff_2x",
		"gpu_utilization_87_percent_memory_allocated_18.4GB_temperature_72C",
		"registry_version_3.2.1_created_2026-01-15_status_active_deprecated_false",
		"prompt_template_system_message_role_user_context_window_128k_sequences",
		"safety_filter_toxicity_threshold_0.85_bias_detection_enabled_content_policy_v2",
		"batch_inference_queue_depth_47_processing_time_avg_156ms_p99_342ms",
		"layer_norm_epsilon_1e-12_hidden_size_3072_intermediate_size_12288_heads_12",
	}

	generator.mutex.Lock()
	defer generator.mutex.Unlock()
	index := generator.randomGen.Intn(len(patterns))
	return patterns[index]
}
