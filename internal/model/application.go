package model

import (
	"encoding/json"
	"time"
)

// Application K8s应用表
type Application struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`            // 主键ID
	Name        string    `gorm:"column:name;not null;type:varchar(255)" json:"name"`         // 应用名称
	Upstream    string    `gorm:"column:upstream;not null;type:varchar(200)" json:"upstream"` // 上游地址
	Description string    `gorm:"column:description;type:text" json:"description"`            // 描述
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"` // 更新时间
}

// TableName 指定表名
func (Application) TableName() string {
	return "applications"
}

// Directory 目录表
type Directory struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                // 自增主键
	AppID     string    `gorm:"column:app_id;not null;type:varchar(64);index" json:"app_id"` // 所属应用ID，外键 references applications
	ParentID  int       `gorm:"column:parent_id;not null;default:0;index" json:"parent_id"`  // 父目录ID，外键 references directories
	Name      string    `gorm:"column:name;not null;type:varchar(255)" json:"name"`          // 目录名称
	SortOrder int       `gorm:"column:sort_order;not null;default:1" json:"sort_order"`      // 排序序号
	CreatedAt time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`  // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`  // 更新时间
}

func (Directory) TableName() string {
	return "directories"
}

// Router 路由表
type Router struct {
	ID             int             `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                          // 自增主键
	AppID          string          `gorm:"column:app_id;not null;type:varchar(64);index" json:"app_id"`           // 所属应用ID
	DirectoryID    int             `gorm:"column:directory_id;not null;default:0;index" json:"directory_id"`      // 所属目录ID
	Path           string          `gorm:"column:path;not null;type:varchar(255)" json:"path"`                    // 路由路径
	Method         string          `gorm:"column:method;not null;default:'GET';type:varchar(10)" json:"method"`   // HTTP方法
	Headers        json.RawMessage `gorm:"column:headers;type:jsonb;default:'[]'" json:"headers"`                 // 请求头限制（JSON数组）
	RequestSchema  json.RawMessage `gorm:"column:request_schema;type:jsonb;default:'{}'" json:"request_schema"`   // 请求体JSON Schema
	ResponseSchema json.RawMessage `gorm:"column:response_schema;type:jsonb;default:'{}'" json:"response_schema"` // 响应体JSON Schema
	CreatedAt      time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`            // 创建时间
	UpdatedAt      time.Time       `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`            // 更新时间
	Descript       string          `gorm:"column:descript;type:text;default:''" json:"descript"`                  // 描述
}

func (Router) TableName() string {
	return "routers"
}

// Workflow 工作流表
type Workflow struct {
	ID           int             `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                           // 自增主键
	AppID        string          `gorm:"column:app_id;not null;type:varchar(64);index:idx_workflows_app_type" json:"app_id"`     // 所属应用ID
	WorkflowType string          `gorm:"column:workflow_type;not null;default:'pre_work';type:varchar(20)" json:"workflow_type"` // 工作流类型：pre_work / post_work
	FuncKey      string          `gorm:"column:func_key;not null;default:'';type:varchar(255)" json:"func_key"`                  // 函数标识
	Params       json.RawMessage `gorm:"column:params;type:jsonb;default:'{}'" json:"params"`                                    // 参数（JSON对象）
	SortOrder    int             `gorm:"column:sort_order;not null;default:1" json:"sort_order"`                                 // 排序序号
	CreatedAt    time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`                             // 创建时间
	UpdatedAt    time.Time       `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`                             // 更新时间
}

func (Workflow) TableName() string {
	return "workflows"
}
