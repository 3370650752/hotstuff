package config

import (
	"slices"
	"time"

	"github.com/relab/hotstuff"
	"github.com/relab/hotstuff/internal/proto/orchestrationpb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ExperimentConfig 保存实验的配置信息。
type ExperimentConfig struct {
	// # 以下是基于主机的配置值

	// ReplicaHosts 是一个主机列表，这些主机将运行副本。
	ReplicaHosts []string
	// ClientHosts 是一个主机列表，这些主机将运行客户端。
	ClientHosts []string
	// Replicas 是副本的总数。
	Replicas int
	// Clients 是客户端的总数。
	Clients int
	// Locations 是一个副本位置列表（可选，但如果设置了 TreePositions 则为必需）。
	// Locations 的长度必须等于副本的数量，但可以包含重复项。
	// 位置通过副本 ID 进行索引。
	// Locations 中的条目必须存在于延迟矩阵中。
	Locations []string
	// ByzantineStrategy 是一个映射，将每个策略映射到表现出该策略的副本 ID 列表。
	ByzantineStrategy map[string][]uint32

	// # 以下是基于树的配置值：

	// TreePositions 是一个副本树位置列表（可选）。
	// TreePositions 的长度必须等于副本的数量，并且条目不重复。
	// 树位置通过副本 ID 进行索引。
	// TreePositions 中的第 0 个条目是树的根，第 1 个条目是根的左子节点，
	// 第 2 个条目是根的右子节点，依此类推。
	TreePositions []uint32
	// BranchFactor 是树的分支因子（如果设置了 TreePositions 则为必需）。
	BranchFactor uint32
	// TreeDelta 是树中中间节点的等待时间。
	TreeDelta time.Duration
	// 如果为 true，则对现有的树位置进行洗牌。
	RandomTree bool

	// # 以下是模块字符串：

	// Consensus 是要使用的共识实现的名称。
	Consensus string
	// Crypto 是要使用的加密实现的名称。
	Crypto string
	// LeaderRotation 是要使用的领导者轮换算法的名称。
	LeaderRotation string
	// Modules 是要加载的其他模块的列表。
	Modules []string
	// Metrics 是要记录的指标列表。
	Metrics []string

	// # 以下是文件路径字符串：

	// Cue 是可选的 .cue 配置文件的路径。
	Cue string
	// Exe 是部署到远程主机的可执行文件的路径。
	Exe string
	// Output 是实验数据输出目录的路径。
	Output string
	// SshConfig 是 SSH 配置文件的路径。
	SshConfig string

	// # 以下是性能分析标志：

	// CpuProfile 如果为 true，则启用 CPU 性能分析。
	CpuProfile bool
	// FgProfProfile 如果为 true，则启用 fgprof 库。
	FgProfProfile bool
	// MemProfile 如果为 true，则启用内存性能分析。
	MemProfile bool

	// DurationSamples 是预测视图持续时间时要考虑的前几个视图的数量。
	DurationSamples uint32

	// # 以下是基于客户端的配置值：

	// MaxConcurrent 是每个客户端并发发送的最大命令数。
	MaxConcurrent uint32
	// PayloadSize 是客户端在一个批次中发送的每个命令的固定大小（以字节为单位）。
	PayloadSize uint32
	// RateLimit 是客户端每秒允许发送的最大命令数。
	RateLimit float64
	// RateStep 是客户端可以提高其速率的每秒命令数。
	RateStep float64
	// 客户端命令应该批量处理的数量。
	BatchSize uint32

	// # 其他值：

	// TimeoutMultiplier 是在发生超时的情况下，将视图持续时间乘以的倍数。
	TimeoutMultiplier float64

	// Worker 在控制器上生成一个本地工作进程。
	Worker bool

	// SharedSeed 是跨节点共享的随机数生成器种子。
	SharedSeed int64
	// Trace 启用运行时跟踪。
	Trace bool
	// LogLevel 是指定日志级别的字符串。
	LogLevel string
	// UseTLS 启用 TLS。
	UseTLS bool

	// # 持续时间值在此处分开，以便于添加对 Cue 的兼容性。
	// 注意：Cue 不支持 time.Duration，必须手动实现。

	// ClientTimeout 指定客户端的超时持续时间。
	ClientTimeout time.Duration
	// ConnectTimeout 指定初始连接超时的持续时间。
	ConnectTimeout time.Duration
	// RateStepInterval 是客户端速率限制应该增加的频率。
	RateStepInterval time.Duration
	// MeasurementInterval 是测量之间的时间间隔。
	MeasurementInterval time.Duration

	// Duration 指定整个实验的持续时间。
	Duration time.Duration
	// ViewTimeout 是第一个视图的持续时间。
	ViewTimeout time.Duration
	// MaxTimeout 是视图超时的上限。
	MaxTimeout time.Duration
}

// TreePosIDs 返回一个按树位置排序的 hotstuff.ID 切片。
func (c *ExperimentConfig) TreePosIDs() []hotstuff.ID {
	ids := make([]hotstuff.ID, 0, len(c.TreePositions))
	for i, id := range c.TreePositions {
		ids[i] = hotstuff.ID(id)
	}
	return ids
}

// ReplicasForHost 返回分配给给定索引处主机的副本数量。
func (c *ExperimentConfig) ReplicasForHost(hostIndex int) int {
	return unitsForHost(hostIndex, c.Replicas, len(c.ReplicaHosts))
}

// ClientsForHost 返回分配给给定索引处主机的客户端数量。
func (c *ExperimentConfig) ClientsForHost(hostIndex int) int {
	return unitsForHost(hostIndex, c.Clients, len(c.ClientHosts))
}

// unitsForHost 返回要分配给 hostIndex 处主机的单元数量。
func unitsForHost(hostIndex int, totalUnits int, numHosts int) int {
	if numHosts == 0 {
		return 0
	}
	unitsPerHost := totalUnits / numHosts
	remainingUnits := totalUnits % numHosts
	if hostIndex < remainingUnits {
		return unitsPerHost + 1
	}
	return unitsPerHost
}

// AssignReplicas 将副本分配给主机。
func (c *ExperimentConfig) AssignReplicas(srcReplicaOpts *orchestrationpb.ReplicaOpts) ReplicaMap {
	hostsToReplicas := make(ReplicaMap)
	nextReplicaID := hotstuff.ID(1)

	for hostIdx, host := range c.ReplicaHosts {
		numReplicas := c.ReplicasForHost(hostIdx)
		for i := 0; i < numReplicas; i++ {
			replicaOpts := srcReplicaOpts.New(nextReplicaID, c.Locations)
			replicaOpts.SetByzantineStrategy(c.lookupByzStrategy(nextReplicaID))
			hostsToReplicas[host] = append(hostsToReplicas[host], replicaOpts)
			nextReplicaID++
		}
	}
	return hostsToReplicas
}

// lookupByzStrategy 返回给定副本的拜占庭策略。
// 如果副本不是拜占庭的，该函数将返回一个空字符串。
// 这假设 replicaID 是有效的；这由 cue 配置解析器检查。
func (c *ExperimentConfig) lookupByzStrategy(replicaID hotstuff.ID) string {
	for strategy, ids := range c.ByzantineStrategy {
		if slices.Contains(ids, uint32(replicaID)) {
			return strategy
		}
	}
	return ""
}

// AssignClients 将客户端分配给主机。
func (c *ExperimentConfig) AssignClients() ClientMap {
	hostsToClients := make(ClientMap)
	nextClientID := hotstuff.ID(1)

	for hostIdx, host := range c.ClientHosts {
		numClients := c.ClientsForHost(hostIdx)
		for i := 0; i < numClients; i++ {
			hostsToClients[host] = append(hostsToClients[host], nextClientID)
			nextClientID++
		}
	}
	return hostsToClients
}

// IsLocal 如果副本和客户端主机切片都只包含一个 "localhost" 实例，则返回 true。请参阅 NewLocal。
func (c *ExperimentConfig) IsLocal() bool {
	if len(c.ClientHosts) > 1 || len(c.ReplicaHosts) > 1 {
		return false
	}
	return (c.ReplicaHosts[0] == "localhost" && c.ClientHosts[0] == "localhost") ||
		(c.ReplicaHosts[0] == "127.0.0.1" && c.ClientHosts[0] == "127.0.0.1")
}

// AllHosts 返回所有主机名的列表，包括副本和客户端。
// 如果配置设置为本地运行，该函数将返回一个包含一个名为 "localhost" 的条目的列表。
func (c *ExperimentConfig) AllHosts() []string {
	if c.IsLocal() {
		return []string{"localhost"}
	}
	return append(c.ReplicaHosts, c.ClientHosts...)
}

// CreateReplicaOpts 根据实验配置创建一个新的 ReplicaOpts。
func (c *ExperimentConfig) CreateReplicaOpts() *orchestrationpb.ReplicaOpts {
	return &orchestrationpb.ReplicaOpts{
		UseTLS:            c.UseTLS,
		BatchSize:         c.BatchSize,
		TimeoutMultiplier: float32(c.TimeoutMultiplier),
		Consensus:         c.Consensus,
		Crypto:            c.Crypto,
		LeaderRotation:    c.LeaderRotation,
		ConnectTimeout:    durationpb.New(c.ConnectTimeout),
		InitialTimeout:    durationpb.New(c.ViewTimeout),
		TimeoutSamples:    c.DurationSamples,
		MaxTimeout:        durationpb.New(c.MaxTimeout),
		SharedSeed:        c.SharedSeed,
		Modules:           c.Modules,
		BranchFactor:      c.BranchFactor,
		TreePositions:     c.TreePositions,
		TreeDelta:         durationpb.New(c.TreeDelta),
	}
}

// CreateClientOpts 根据实验配置创建一个新的 ClientOpts。
func (c *ExperimentConfig) CreateClientOpts() *orchestrationpb.ClientOpts {
	return &orchestrationpb.ClientOpts{
		UseTLS:           c.UseTLS,
		ConnectTimeout:   durationpb.New(c.ConnectTimeout),
		PayloadSize:      c.PayloadSize,
		MaxConcurrent:    c.MaxConcurrent,
		RateLimit:        c.RateLimit,
		RateStep:         c.RateStep,
		RateStepInterval: durationpb.New(c.RateStepInterval),
		Timeout:          durationpb.New(c.ClientTimeout),
	}
}

// ReplicaMap 将主机映射到副本选项切片。
type ReplicaMap map[string][]*orchestrationpb.ReplicaOpts

// ReplicaIDs 返回在给定主机上运行的副本的 ID。
func (r ReplicaMap) ReplicaIDs(host string) []uint32 {
	ids := make([]uint32, 0, len(r[host]))
	for _, opts := range r[host] {
		ids = append(ids, opts.ID)
	}
	return ids
}

// ClientMap 将主机映射到客户端 ID 切片。
type ClientMap map[string][]hotstuff.ID

// ClientIDs 返回在给定主机上运行的客户端的 ID。
func (c ClientMap) ClientIDs(host string) []uint32 {
	ids := make([]uint32, 0, len(c[host]))
	for _, id := range c[host] {
		ids = append(ids, uint32(id))
	}
	return ids
}
