<template>
    <div class="am-host-container">
        <div class="am-host-container__operator">
            <el-card>
                <el-select v-model="timeDensity" placeholder="Select" size="default" style="width: 240px">
                    <el-option v-for="item in options" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
            </el-card>
        </div>
        <el-row :gutter="4">
            <el-col :span="12">
                <el-card>
                    <echarts :option="cpuOption">
                        <div class="am-host-container__image-title">CPU 总使用率</div>
                        <div class="am-host-container__image-description">百分比： {{ cpuPercent }}</div>
                    </echarts>
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card>
                    <echarts :option="memOption">
                        <div class="am-host-container__image-title">内存使用率</div>
                        <div class="am-host-container__image-description">
                            总量：{{ memInfo.total }} 使用：{{ memInfo.used }} 百分比： {{ memInfo.percent }}
                        </div>
                    </echarts>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="4">
            <el-col :span="12" v-for="(item) in netOptionList" :key="item.sourceInfo.ethernet">
                <el-card>
                    <echarts :option="item">
                        <div class="am-host-container__image-title">流量曲线图</div>
                        <div class="am-host-container__image-description">
                            {{ item.sourceInfo.ethernet }} 接收：{{ item.sourceInfo.read }} 发送：{{ item.sourceInfo.write }}
                        </div>
                    </echarts>
                </el-card>
            </el-col>
            <el-col :span="12" v-for="(item, index) in diskOptionList" :key="index">
                <el-card>
                    <echarts :option="item">
                        <div class="am-host-container__image-title">磁盘使用率</div>
                        <div class="am-host-container__image-description">
                           目录：{{ (item.sourceInfo as DiskUsage).mountpoint }} 总量：{{ (item.sourceInfo as DiskUsage).total }} 使用：{{ (item.sourceInfo as DiskUsage).used }} 百分比：{{ (item.sourceInfo as DiskUsage).percent }}
                        </div>
                    </echarts>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>
<script setup lang="ts">
import {
    queryCPUInfo,
    queryCPUUsage,
    queryDiskInfo,
    queryDiskUsage,
    queryMemInfo,
    queryMemUsage,
    queryNetworkUsage
} from '@/api/host'
import { EChartsOption } from '@/components/Echarts/echarts.ts'
import { cpuOptions, diskOptions, memOptions, netOptions } from '@/components/Echarts/line.ts'
import { CPUTrendingArgs, DiskIO, DiskTrendingArgs, DiskUsage, MemTrendingArgs, NetIO, NetTrendingArgs } from '@/interface/host.ts'
import { convertBytesToReadable } from '@/utils/convert.ts'
import { dayjs } from 'element-plus'
import { set } from 'lodash-es'

// 时间密度下拉框
const timeDensity = ref(43200)
const options = [
    {
        value: 600,
        label: '10分钟'
    },
    {
        value: 1800,
        label: '30分钟'
    },
    {
        value: 3600,
        label: '1 小时'
    },
    {
        value: 43200,
        label: '12小时'
    },
    {
        value: 86400,
        label: '24小时'
    }
]

const cpuPercent = ref('0.0%')
const renderCPUPercent = async () => {
    const { data } = await queryCPUInfo()
    cpuPercent.value = data.percent.toFixed(2) + '%'
}

const cpuOption = reactive<EChartsOption>(cpuOptions)
const renderCPU = async () => {
    const param: CPUTrendingArgs = {
        start_time: dayjs().unix() - timeDensity.value,
        end_time: dayjs().unix()
    }
    console.log(param)
    const { data } = await queryCPUUsage(param)
    const cpuData = data.data
    console.log('cpu response:', cpuData)
    // set(cpuOption, 'title', { text: 'CPU使用率' });
    set(
        cpuOption,
        'xAxis.data',
        cpuData.map((item) => dayjs(item.timestamp * 1000).hour() + ':' + dayjs(item.timestamp * 1000).minute())
    )
    set(cpuOption, 'legend.data', ['CPU使用率'])
    set(cpuOption, 'series', [
        {
            name: 'CPU使用率',
            data: cpuData.map((item) => item.value),
            type: 'line',
            smooth: true,
            showSymbol: false
        }
    ])
    console.log('cpu: ', cpuOption)
}

const memInfo = ref({
    percent: '0%',
    total: '0',
    used: '0'
})

const renderMemInfo = async () => {
    const { data } = await queryMemInfo()
    memInfo.value.percent = data.percent.toFixed(2) + '%'
    memInfo.value.total = convertBytesToReadable(data.total)
    memInfo.value.used = convertBytesToReadable(data.used)
}

const memOption = reactive<EChartsOption>(memOptions) as EChartsOption
const renderMem = async () => {
    const param: MemTrendingArgs = {
        start_time: dayjs().unix() - timeDensity.value,
        end_time: dayjs().unix()
    }
    console.log(param)
    const { data } = await queryMemUsage(param)
    const memData = data.data
    console.log('mem response: ', memData)
    // set(memOption, 'title', { text: '内存使用率' });
    set(
        memOption,
        'xAxis.data',
        memData.map((item) => dayjs(item.timestamp * 1000).hour() + ':' + dayjs(item.timestamp * 1000).minute())
    )
    set(memOption, 'legend.data', ['内存使用率'])
    set(memOption, 'series', [
        {
            name: '内存使用率',
            data: memData.map((item) => item.value),
            type: 'line',
            smooth: true,
            showSymbol: false
        }
    ])
    console.log('mem: ', memOption)
}

const diskInfo = ref<
    {
        device: string
        total: string
        used: string
        percent: string
    }[]
>([])

const renderDiskInfo = async () => {
    const { data } = await queryDiskInfo()
    console.log(data)
    diskInfo.value = []
    data.info.map((item) => {
        diskInfo.value.push({
            device: item.device,
            total: convertBytesToReadable(item.total),
            used: convertBytesToReadable(item.used),
            percent: item.percent.toFixed(2) + '%'
        })
    })
}

const diskOptionList = ref(<EChartsOption[]>([]) as EChartsOption[])
const renderDisk = async () => {
    const param: DiskTrendingArgs = {
        start_time: dayjs().unix() - timeDensity.value,
        end_time: dayjs().unix()
    }
    console.log(param)
    const { data } = await queryDiskUsage(param)
    const diskData = data
    console.log('disk response: ', diskData)
    const newDiskOptionList = diskData.map((item: DiskUsage) => {
        const newOption: EChartsOption = {
            ...diskOptions,
            sourceInfo: {
                device: item.device,
                total: convertBytesToReadable(item.total),
                used: convertBytesToReadable(item.used),
                percent: item.percent.toFixed(2) + '%',
                mountpoint: item.mountpoint
            },
            xAxis: {
                data: item.data.map(
                    (item: DiskIO) => dayjs(item.timestamp * 1000).hour() + ':' + dayjs(item.timestamp * 1000).minute()
                )
            },
            legend: {
                data: ['Read', 'Write']
            },
            series: [
                {
                    name: 'Read',
                    data: item.data.map((item: DiskIO) => item.io_read),
                    type: 'line',
                    smooth: true,
                    showSymbol: false
                },
                {
                    name: 'Write',
                    data: item.data.map((item: DiskIO) => item.io_write),
                    type: 'line',
                    smooth: true,
                    showSymbol: false
                }
            ]
        }
        return newOption
    })
    diskOptionList.value = newDiskOptionList
}

const netOptionList = ref(<EChartsOption[]>([]) as EChartsOption[])
const renderNet = async () => {
    const param: NetTrendingArgs = {
        start_time: dayjs().unix() - timeDensity.value,
        end_time: dayjs().unix()
    }
    console.log(param)
    const { data } = await queryNetworkUsage(param)
    const netData = data
    console.log('net response: ', netData)
    const newNetOptionList = netData.map((item) => {
        // sum up the bytes_recv and bytes_sent
        var totalRecv =  item.data.reduce((acc: number, cur: NetIO) => acc + cur.bytes_recv, 0)
        var totalSent =  item.data.reduce((acc: number, cur: NetIO) => acc + cur.bytes_sent, 0)
        const netOption: EChartsOption = {
            ...netOptions,
            sourceInfo: {
                ethernet: item.ethernet,
                read: convertBytesToReadable(totalRecv),
                write: convertBytesToReadable(totalSent)
            },
            xAxis: {
                data: item.data.map(
                    (item: NetIO) => dayjs(item.timestamp * 1000).hour() + ':' + dayjs(item.timestamp * 1000).minute()
                )
            },
            legend: {
                data: ['Receive', 'Send']
            },
            series: [
                {
                    name: 'Receive',
                    data: item.data.map((item: NetIO) => item.bytes_recv),
                    type: 'line',
                    smooth: true,
                    showSymbol: false
                },
                {
                    name: 'Send',
                    data: item.data.map((item: NetIO) => item.bytes_sent),
                    type: 'line',
                    smooth: true,
                    showSymbol: false
                }
            ]
        }
        return netOption
    })
    netOptionList.value = newNetOptionList
}
const timer = ref()
onMounted(() => {
    console.log('mounted')
    renderCPUPercent()
    renderCPU()
    renderMemInfo()
    renderMem()
    renderDisk()
    renderNet()
    timer.value = setInterval(() => {
        console.log('start interval')
        renderCPUPercent()
        renderCPU()
        renderMemInfo()
        renderMem()
        renderDisk()
        renderNet()
    }, 5000)
    console.log('timer: ', timer.value)
})

onUnmounted(() => {
    console.log('unmounted')
    clearInterval(timer.value)
})

watch(
    () => timeDensity.value,
    () => {
        renderCPUPercent()
        renderCPU()
        renderMemInfo()
        renderMem()
        renderDiskInfo()
        renderDisk()
        renderNet()
    }
)
</script>

<style scoped lang="scss">
@include b(host-container) {
    overflow: scroll;
    height: 100%;
    background-color: #ffffff;
    .el-row {
        margin-bottom: 4px;
        .el-col {
            height: 310px;
        }
    }

    .el-card {
        height: 100%;
        width: 100%;
        :deep(.el-card__body) {
            height: 100% !important;
            width: 100% !important;
        }
    }

    @include e(operator) {
        height: 48px;
        width: 100%;
        margin-bottom: 4px;
        .el-card {
            height: 100%;
            :deep(.el-card__body) {
                height: 100% !important;
                padding: 0 8px 0 0;
                display: flex;
                align-items: center;
                justify-content: flex-end;
            }
        }
    }

    @include e(image-title) {
        font-size: 14px;
        font-weight: bold;
    }

    @include e(image-description) {
        font-size: 12px;
        color: #73767a;
    }
}
</style>
