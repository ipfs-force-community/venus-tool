
import Card from "@/component/card"
import { Descriptions, Empty } from "antd"
import { useParams } from "react-router-dom"
import { useDealInfo } from "@/fetcher"


export default function (props) {
    const params = useParams()
    let { id, data } = props
    const title = "Deal Info"
    // check data
    if (!id) {
        id = params.id
    }

    const { data: deal, isLoading } = useDealInfo(id)

    if (!data) {
        data = deal
    }

    if (isLoading) {
        return (<Empty />)
    }

    return (
        <>
            <Card title={title}>
                <Descriptions
                    column={1}
                    bordered
                    size="small"
                >
                    <Descriptions.Item label="ProposalCid">{data.ProposalCid["/"]}</Descriptions.Item>
                    <Descriptions.Item label="PieceCID">{data.Proposal.PieceCID["/"]}</Descriptions.Item>
                    <Descriptions.Item label="Deal ID">{data.DealID}</Descriptions.Item>
                    <Descriptions.Item label="State">{data.State}</Descriptions.Item>
                    <Descriptions.Item label="PieceStatus">{data.PieceStatus}</Descriptions.Item>
                    <Descriptions.Item label="Provider">{data.Proposal.Provider}</Descriptions.Item>
                    <Descriptions.Item label="Client">{data.Proposal.Client}</Descriptions.Item>
                    <Descriptions.Item label="VerifiedDeal">{data.Proposal.VerifiedDeal ? "True" : "False"}</Descriptions.Item>
                    <Descriptions.Item label="StartEpoch">{data.Proposal.StartEpoch}</Descriptions.Item>
                    <Descriptions.Item label="EndEpoch">{data.Proposal.EndEpoch}</Descriptions.Item>
                    <Descriptions.Item label="ProviderCollateral">{data.Proposal.ProviderCollateral}</Descriptions.Item>
                    <Descriptions.Item label="ClientCollateral">{data.Proposal.ClientCollateral}</Descriptions.Item>
                    <Descriptions.Item label="PayloadSize">{data.PayloadSize}</Descriptions.Item>
                    <Descriptions.Item label="PieceSize">{data.Proposal.PieceSize}</Descriptions.Item>
                    <Descriptions.Item label="Offset">{data.Offset}</Descriptions.Item>
                    <Descriptions.Item label="SectorNumber">{data.SectorNumber}</Descriptions.Item>
                    <Descriptions.Item label="StoragePricePerEpoch">{data.Proposal.StoragePricePerEpoch}</Descriptions.Item>
                </Descriptions>
            </Card>
        </>
    )
}
