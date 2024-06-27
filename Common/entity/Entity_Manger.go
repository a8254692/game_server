package entity

import (
	"BilliardServer/Util/log"
	"sync"
)

// 实体数据管理器
type Entity_Manager struct {
	EntityMap         map[uint32]Entity //实体列表
	EntityCount       int               //实体数量
	EntityManagerName string            //管理器名称

	lock sync.RWMutex
}

func (this *Entity_Manager) Init(name string) {
	this.EntityMap = make(map[uint32]Entity)
	this.EntityCount = 0
	this.EntityManagerName = name
	log.Info("-->实体管理器初始化完成 ", name)
}

// 依据EntityID获取Entity
func (this *Entity_Manager) GetEntityByID(tEntityID uint32) Entity {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.EntityMap[tEntityID]
}

func (this *Entity_Manager) Contain(id uint32) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.EntityMap[id] != nil
}

// 通过ObjID获取Entity对象
func (this *Entity_Manager) GetEntityByObjID(objID string) Entity {
	this.lock.RLock()
	defer this.lock.RUnlock()

	for _, entity := range this.EntityMap {
		if entity.GetObjID() == objID {
			//相同返回
			return entity
		}
	}
	return nil

}

// 删除Entity
func (this *Entity_Manager) DelEntity(tEntity Entity) {
	this.lock.Lock()
	defer this.lock.Unlock()

	delete(this.EntityMap, tEntity.GetEntityID())
	this.EntityCount = len(this.EntityMap)
}

// 添加Entity
func (this *Entity_Manager) AddEntity(tEntity Entity) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.EntityMap[tEntity.GetEntityID()] = tEntity
	this.EntityCount = len(this.EntityMap)
}

// 所有Entity执行
func (this *Entity_Manager) AllEntityDoFunc(f func(tEntity Entity)) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, tEntity := range this.EntityMap {
		f(tEntity)
	}
}
