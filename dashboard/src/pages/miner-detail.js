import Card from "@/component/card"
import { Descriptions, Space } from "antd"
import { gg, useParams } from "react-router-dom"
import SectorList from "../component/sector-list"
import { useMinerInfo, rel } from "../fetcher"
import useSWR from "swr"
import { Fil, Power } from "../util"


export default function ({ id }) {
    const params = useParams()
    // const navigate = useNavigate()

    const title = "Miner Info"
    // check data
    if (!id) {
        id = params.id
    }

    const { data, error } = useMinerInfo(id)
    if (error) {
        throw error
    }
    if (!data) {
        return (
            <>
                <Card title={title} loading={true} />
            </>
        )
    }

    return (
        <>
            <Space style={{ width: '100%' }} direction='vertical' size={'large'}>
                <Card title={title}>
                    <Descriptions
                        column={1}
                        bordered
                        size="small"
                    >
                        <Descriptions.Item label="Owner">{data.Owner}</Descriptions.Item>
                        <Descriptions.Item label="Beneficiary">{data.Beneficiary}</Descriptions.Item>
                        <Descriptions.Item label="Worker">{data.Worker}</Descriptions.Item>
                        {data.NewWorker === "<empty>" ? null : (<Descriptions.Item label="NewWorker">{data.NewWorker}</Descriptions.Item>)}
                        {data.NewWorker === "<empty>" ? null : (<Descriptions.Item label="WorkerChangeEpoch">{data.WorkerChangeEpoch}</Descriptions.Item>)}

                        <Descriptions.Item label="ControlAddresses">{data.ControlAddresses}</Descriptions.Item>
                        <Descriptions.Item label="AvailBalance">{Fil(data.AvailBalance)}</Descriptions.Item>
                        <Descriptions.Item label="PeerId">{data.PeerId}</Descriptions.Item>
                        <Descriptions.Item label="Multiaddrs">{data.Multiaddrs}</Descriptions.Item>
                        <Descriptions.Item label="SectorSize">{data.SectorSize}</Descriptions.Item>
                        <Descriptions.Item label="WindowPoStProofType">{data.WindowPoStProofType}</Descriptions.Item>
                        <Descriptions.Item label="windowPoStPartitionSectors">{data.WindowPoStPartitionSectors}</Descriptions.Item>
                        <Descriptions.Item label="RawBytePower">{`${Power(data.MinerPower.RawBytePower)} (${inPercent(data.MinerPower.RawBytePower, data.TotalPower.RawBytePower)})`}</Descriptions.Item>
                        <Descriptions.Item label="QualityAdjPower">{`${Power(data.MinerPower.QualityAdjPower)} (${inPercent(data.MinerPower.QualityAdjPower, data.TotalPower.QualityAdjPower)})`}</Descriptions.Item>

                        {/* <Descriptions.Item label="Deadline">{data.Deadline}</Descriptions.Item> */}
                    </Descriptions>
                </Card>
                <SectorList miner={id} />
            </Space>
        </>
    )
}

const inPercent = (a, b) => {
    a = Number(a)
    b = Number(b)

    if (b === 0) {
        return "0%"
    }
    const p = a / b
    return (p * 100).toFixed(2) + "%"
}
