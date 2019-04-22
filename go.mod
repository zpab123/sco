module github.com/zpab123/sco

require (
	"github.com/zpab123/zaplog"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
	"github.com/gorilla/websocket"
	"github.com/go-ini/ini"
	"github.com/zpab123/syncutil"
	"golang.org/x/net/context"
	"github.com/gogo/protobuf/proto"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

go 1.12
