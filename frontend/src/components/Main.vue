<script setup>
import { ref, watch, computed} from 'vue'
import { InitNode,GetInfo,GetAllNode,GetBestNode,StartTest,PingSelectedNode,GetSpeed } from '../../wailsjs/go/service/App'
import { ElMessage } from 'element-plus'

const userInfo = ref({})
const bestInfo = ref({})
const nodeInfo = ref([])
const rawNodeList = ref([])
const selectedNodeIds = ref([])
const selectedNode = ref({})
const multiSelectEnabled = ref(false)
const speedTestMode = ref('0')
const threads = ref(4)
const downSpeed = ref(0);
const upSpeed = ref(0);
const downData = ref(0);
const upData = ref(0);
const btnDis = ref(0)

const isDisabled = computed(() => {
  return btnDis.value === 1
})

async function main() {
  btnDis.value = 1
  if (selectedNodeIds.value.length != 0) {
    selectedNodeIds.value = []
  } 
  const res = await asyInitNode();
  if (res.success) {
   const res = await asyGetInfo();
    if (res.success) {
      const res = await asyAllNode();
      if (res.success) {
        asyParseNode()
      } else {
        console.log("数据加载失败")
      }
    } else {
      console.log("数据加载失败")
    }
  } else {
    console.log("数据加载失败")
  }
}

main()

watch(
    () => selectedNode.value,
    () => {
        selectedNode.value.NewBandWidth = ParseBandWidth(selectedNode.value.BandWidth)
        selectedNode.value.Latency = "测试中..."
        var idx
        if (!selectedNode.value.id) {
          idx = -1
        } else {
          idx = selectedNode.value.id
        }
        console.log(idx)
        PingSelectedNode(idx).then((result) => {
          selectedNode.value.Latency = result
        })
    }
);

watch(
  () => multiSelectEnabled.value,
  () => {
    handleNodeChange()
  }
);

function handleNodeChange() {
  if (!multiSelectEnabled.value && selectedNodeIds.value.length > 1) {
    selectedNodeIds.value = [selectedNodeIds.value[selectedNodeIds.value.length - 1]]
  } 
   if (nodeInfo.value.value.length > 0 && selectedNodeIds.value.length > 0) {
    const idx = selectedNodeIds.value[0]
    selectedNode.value = nodeInfo.value.value[idx]
    console.log(selectedNode.value);
  } else {
    selectedNode.value = bestInfo.value
  }
}


function startTest() {
  btnDis.value = 1
  if(multiSelectEnabled.value){
    threads.value = 1
    ElMessage.warning("多节点模式下不支持设置线程")
  }
  if (!selectedNodeIds.value.length) {
    ElMessage.success(`测速开始！默认使用最佳节点, 线程数:${threads.value}, 模式:${speedTestMode.value}`)
    StartTest([], threads.value, parseInt(speedTestMode.value))
    fetchSpeedData()
  } else {
    ElMessage.success(`测速开始！${selectedNode.value.Name}, 线程数:${threads.value}, 模式:${speedTestMode.value}`)
    StartTest(selectedNodeIds.value, threads.value, parseInt(speedTestMode.value))
    fetchSpeedData()
  }
}

function ParseBandWidth(value){
    const units = ['bps','Kbps', 'Mbps', 'Gbps', 'Tbps'];
    let i = 0;
    value = value * 8
    while (value >= 1000 && i < units.length - 1) {
      value /= 1000;
      i++;
    }
  let formattedValue;
  if (value >= 100) {
    formattedValue = Math.round(value);
  } else if (value >= 10) {
    formattedValue = value.toFixed(1);
  } else {
    formattedValue = value.toFixed(2);
  }
  return `${formattedValue} ${units[i]}`;
}

function ParseData(value){
    const units = ['Byte','KB', 'MB', 'GB', 'TB'];
    let i = 0;
    while (value >= 1024 && i < units.length - 1) {
      value /= 1024;
      i++;
    }
  let formattedValue;
  if (value >= 100) {
    formattedValue = Math.round(value);
  } else if (value >= 10) {
    formattedValue = value.toFixed(1);
  } else {
    formattedValue = value.toFixed(2);
  }
  return `${formattedValue} ${units[i]}`;
}

async function fetchSpeedData() {
  return new Promise(resolve => {
    setTimeout(() => {
      GetSpeed().then((result) =>{
        try {
          const parsedData = JSON.parse(result)
            downSpeed.value = parsedData.DownSpeedKBps * 1024
            upSpeed.value = parsedData.UpSpeedKBps * 1024
            downData.value = parsedData.TotalDData 
            upData.value = parsedData.TotalUData
          resolve({ success: true})
          if (parsedData.Is_done != 1){
            const res = fetchSpeedData();
          } else {
            btnDis.value = 0
            downSpeed.value = parsedData.DownSpeedKBps * 1024
            upSpeed.value = parsedData.UpSpeedKBps * 1024
            downData.value = parsedData.TotalDData 
            upData.value = parsedData.TotalUData
          }
        } catch (error) {
          btnDis.value = 0
          console.error('JSON 解析失败:', error)
          ElMessage.error("JSON 解析失败!")
          resolve({ success: false})
        } 
      })
    }, 1000)
  })
}

async function asyInitNode() {
  return new Promise(resolve => {
    setTimeout(() => {
      InitNode().then((result) =>{
        if(result == "OK"){
          resolve({ success: true})
        } else {
          ElMessage.error(result)
          resolve({ success: false})
        }
      })
    }, 1000)
  })
}

async function asyAllNode() {
  return new Promise(resolve => {
    setTimeout(() => {
       GetAllNode()
    .then((result) => {
      try {
        const parsedData = JSON.parse(result)
          rawNodeList.value = parsedData
          GetBestNode()
            .then((result) => {
              try {
              const parsedData = JSON.parse(result)
                bestInfo.value = parsedData
                ElMessage.info('获取节点信息成功')
              } catch (error) {
                console.error('JSON 解析失败:', error)
                ElMessage.warning("获取最佳节点失败\n默认使用第一个")
                bestInfo.value = userInfo[0]
              }
            })
            resolve({ success: true})
        } catch (error) {
          console.error('JSON 解析失败:', error)
          ElMessage.error("JSON 解析失败!")
          resolve({ success: false})
        } 
    })
    }, 1000)
  })
}

async function asyGetInfo(){
  return new Promise(resolve => {
    setTimeout(() => {
      GetInfo()
    .then((result) => {
      try {
        const parsedData = JSON.parse(result)
        userInfo.value = parsedData
        if(userInfo.value.HostIP != undefined && userInfo.value.HostIP != ""){
          ElMessage.info('获取用户信息成功')
        } else {
          ElMessage.error('获取用户信息失败')
        }
        resolve({ success: true})
      } catch (error) {
          console.error('JSON 解析失败:', error)
          ElMessage.error("JSON 解析失败!")
          resolve({ success: false})
        }
      })
    }, 1000)
  })
  
}

async function asyParseNode() {
   return new Promise(resolve => {
    setTimeout(() => {
      rawNodeList.value.forEach((item, index) => {
        item.id = index;
        item.label = "节点: "+item.Name+"-"+item.Description
      });
      selectedNode.value = bestInfo.value
      nodeInfo.value = rawNodeList
      console.log(nodeInfo);
      btnDis.value = 0
      resolve({ success: true})
    }, 1000)
  })
}

</script>

<template>
  <el-container style="height: 100vh;">
    <el-aside width="250px" style="background-color: #f5f7fa; padding: 20px;">
      <!-- 侧边栏：UserInfo -->
      <h3>用户信息</h3>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="IP地址">{{ userInfo.HostIP }}</el-descriptions-item>
        <el-descriptions-item label="国家">{{ userInfo.Country }}</el-descriptions-item>
        <el-descriptions-item label="省份">{{ userInfo.Province }}</el-descriptions-item>
        <el-descriptions-item label="市区">{{ userInfo.City + userInfo.District }}</el-descriptions-item>
        <el-descriptions-item label="运营商">{{ userInfo.ISP }}</el-descriptions-item>
      </el-descriptions>
    <!-- 侧边栏：nodeInfo -->
      <h3>节点信息</h3>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="名称">{{ selectedNode.Name }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{ selectedNode.Description }}</el-descriptions-item>
          <el-descriptions-item label="节点带宽">{{ selectedNode.NewBandWidth }}</el-descriptions-item>
          <el-descriptions-item label="节点延迟">{{ selectedNode.Latency }}</el-descriptions-item>
      </el-descriptions>
    </el-aside>
    <!-- 主界面 -->
    <el-container>
      <el-main style="margin: 0 20px 20px 20px; background-color: #fff; padding: 20px;">
      <!-- 速度信息区域 -->
        <div class="speed-display">
        <!-- 下载速度 -->
        <el-card class="speed-card download">
          <div class="speed-direction">
            <el-icon name="download" />
            <span>下载速度</span>
          </div>
          <div class="speed-value">
            {{ ParseBandWidth(downSpeed) }}
          </div>
          <div class="speed-direction">
            <el-icon name="download" />
            <span>使用流量</span>
          </div>
          <div class="data-value">
            {{ ParseData(downData) }}
          </div>  
        </el-card>
        <!-- 上传速度 -->
        <el-card class="speed-card upload">
          <div class="speed-direction">
            <el-icon name="upload" />
            <span>上传速度</span>
          </div>
          <div class="speed-value">
            {{ ParseBandWidth(upSpeed) }}
          </div>
          <div class="speed-direction">
            <el-icon name="upload" />
            <span>使用流量</span>
          </div>
          <div class="data-value">
            {{ ParseData(upData) }}
          </div>
        </el-card>
      </div>
      <br />
      <br />
      <br />
        <!-- 选择区域 -->
        <el-row :gutter="20" style="margin-bottom: 20px;">
          <el-col :span="6">
            <el-select v-model="selectedNodeIds"
                       placeholder="请选择节点"
                       multiple
                       clearable
                       @change="handleNodeChange">
              <!-- 默认显示最佳节点选项 -->
              <el-option
                v-for="node in nodeInfo.value"
                :key="node.id"
                :value="node.id"
                :label="`${node.label}`"
              />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-radio-group v-model="speedTestMode" size="small">
              <el-radio-button label="0">完整</el-radio-button>
              <el-radio-button label="1">仅下载</el-radio-button>
              <el-radio-button label="2">仅上传</el-radio-button>
            </el-radio-group>
          </el-col>
          <el-col :span="4">
            <el-checkbox v-model="multiSelectEnabled">
              开启多选
            </el-checkbox>
          </el-col>
          <el-slider v-model="threads" :min="1" :max="64" show-input style="width: 90%;" />
        </el-row>
        <!-- 按钮区域 -->
        <el-row style="text-align: right;">
          <el-button :disabled=isDisabled type="primary" @click="startTest()">开始测速</el-button>
          <el-button :disabled=isDisabled type="primary" @click="main()">刷新节点</el-button>
        </el-row>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.speed-display {
  display: flex;
  justify-content: space-around;
  padding: 20px 0;
  gap: 20px;
}
.speed-card {
  flex: 1;
  text-align: center;
  border-radius: 12px;
  padding: 20px;
  max-width: 300px;
}
.speed-direction {
  margin-bottom: 10px;
  font-size: 16px;
  font-weight: bold;
  display: flex;
  align-items: center;
  gap: 8px;
}
.speed-value {
  font-size: 2rem;
  font-weight: bold;
  color: #409EFF;
}
.download .speed-value {
  color: #409EFF;
}
.upload .speed-value {
  color: #F56C6C;
}
.el-card.download {
  border-left: 5px solid #409EFF;
}
.el-card.upload {
  border-left: 5px solid #F56C6C;
}

h3 {
  margin: 10px 0;
}
</style>