
import Card from "@/component/card"
import { Descriptions, Space } from "antd"
import { useParams } from "react-router-dom"
import MessageList from "@/component/msg-list"
import { useWalletInfo } from "../fetcher"
import { Editable } from "../component/util"
import { Fil } from "../util.ts"
import axios from "axios"
import { ApiBase } from "../global"


export default function WalletDetail({ id }) {
    const params = useParams()

    const title = "Wallet Info"
    if (!id) {
        id = params.id
    }
    const { data, mutate } = useWalletInfo(id)
    if (!data) {
        return (
            <>
                <Card title={title} loading={true} />
            </>
        )
    }

    const update = (data) => {
        axios.put(ApiBase + "/addr/operate", data).then(() => {
            mutate()
        })
    }

    const updateBaseFee = (newValue) => {
        let body = data
        body.baseFeeStr = newValue
        body.Operate = "set"
        body.IsSetSpec = true

        update(body)
    }
    const updateMaxFee = (newValue) => {
        let body = data
        body.maxFeeStr = newValue
        body.Operate = "set"
        body.IsSetSpec = true
        update(body)
    }

    const updateGasFeeCap = (newValue) => {
        let body = data
        body.gasFeeCapStr = newValue
        body.Operate = "set"
        body.IsSetSpec = true
        update(body)
    }

    const updateGasOverEstimation = (newValue) => {
        let body = data
        body.gasOverEstimation = parseFloat(newValue)
        body.Operate = "set"
        body.IsSetSpec = true

        update(body)
    }


    const updateGasOverPremium = (newValue) => {
        let body = data
        body.gasOverPremium = parseFloat(newValue)
        body.Operate = "set"
        body.IsSetSpec = true

        update(body)
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
                        <Descriptions.Item label="Address">{data.Address}</Descriptions.Item>
                        <Descriptions.Item label="Balance">{Fil(data.Balance ? data.Balance : 0)}</Descriptions.Item>
                        <Descriptions.Item label="Nonce">{data.Nonce}</Descriptions.Item>
                        <Descriptions.Item label="Pending">{data.nonce - data.Nonce}</Descriptions.Item>
                        <Descriptions.Item label="Base Fee"><Editable text={data.baseFee} onUpdate={updateBaseFee} /></Descriptions.Item>
                        <Descriptions.Item label="Max Fee">
                            <Editable text={data.maxFee} onUpdate={updateMaxFee} />
                        </Descriptions.Item>
                        <Descriptions.Item label="Gas Over Estimation">
                            <Editable text={data.gasOverEstimation} onUpdate={updateGasOverEstimation} />
                        </Descriptions.Item>
                        <Descriptions.Item label="Gas Over Premium">
                            <Editable text={data.gasOverPremium} onUpdate={updateGasOverPremium} />
                        </Descriptions.Item>
                        <Descriptions.Item label="Gas Fee Cap">
                            <Editable text={data.gasFeeCap} onUpdate={updateGasFeeCap} />
                        </Descriptions.Item>
                    </Descriptions>
                </Card>
                <MessageList pageSize={100} wallets={[id]} />
            </Space>
        </>
    )
}
