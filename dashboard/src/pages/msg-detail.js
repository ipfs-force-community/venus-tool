import Card from "@/component/card"
import { Descriptions, Empty, Typography } from "antd"
import { useParams } from "react-router-dom"
import { useMsgInfo } from "../fetcher"
import { Copyable } from "../component/util"

export default function (props) {
    const params = useParams()
    let { id, data } = props

    if (!id) {
        id = params.id
    }

    const { data: msg, isLoading } = useMsgInfo(id)

    if (!data) {
        data = msg
    }
    if (isLoading) {
        return (<Empty />)
    }

    console.log("json:", data, data.ParamsInJson, JSON.stringify(data.ParamsInJson));

    return (
        <>
            <Card title={"Message Info"}>
                <Descriptions
                    column={1}
                    bordered
                    size="small"
                >
                    <Descriptions.Item label="ID">{data.ID}</Descriptions.Item>
                    <Descriptions.Item label="SignedCid"><Copyable>{data.SignedCid ? data.SignedCid["/"] : null}</Copyable></Descriptions.Item>
                    <Descriptions.Item label="From">{data.From}</Descriptions.Item>
                    <Descriptions.Item label="To">{data.To}</Descriptions.Item>
                    <Descriptions.Item label="Value">{data.Value}</Descriptions.Item>
                    <Descriptions.Item label="Method">{data.MethodName ? data.MethodName : data.Method}</Descriptions.Item>
                    <Descriptions.Item label="Params"><Copyable>{data.Params}</Copyable></Descriptions.Item>
                    <Descriptions.Item label="Nonce">{data.Nonce}</Descriptions.Item>
                    <Descriptions.Item label="GasLimit">{data.GasLimit}</Descriptions.Item>
                    <Descriptions.Item label="GasFeeCap">{data.GasFeeCap}</Descriptions.Item>
                    <Descriptions.Item label="GasPremium">{data.GasPremium}</Descriptions.Item>
                    <Descriptions.Item label="Height">{data.Height}</Descriptions.Item>
                    <Descriptions.Item label="GasUsed">{data.Receipt?.GasUsed}</Descriptions.Item>
                    <Descriptions.Item label="ExitCode">{data.Receipt?.ExitCode}</Descriptions.Item>
                    <Descriptions.Item label="Return">{data.Receipt?.Return}</Descriptions.Item>
                    <Descriptions.Item label="ErrorMsg">{data.ErrorMsg}</Descriptions.Item>
                    <Descriptions.Item label="Decoded Params"><Copyable><Typography.Text>{JSON.stringify(data.ParamsInJson, null, 4)}</Typography.Text></Copyable></Descriptions.Item>
                </Descriptions>
            </Card >
        </>
    )
}
