import { Col, Row, Table, Popover, Space, Descriptions, Button } from "antd"
import { InfoCircleOutlined, FormOutlined } from '@ant-design/icons';
import { useState } from "react"
import { getDefaultFilters, InShort } from "./util";
import Card from "./card";
import { useDeals, useMiners } from "../fetcher";
import { dealState } from "../util";

export default function DealList(props) {
    const [selectedRowKeys, setSelectedRowKeys] = useState([]);
    const { data: miners } = useMiners()
    const { data: deals } = useDeals({ miner: miners[0] })

    const ret = function (content) {
        return (
            <Card title={"Deals"} >
                {content}
            </Card>
        )
    }

    const preprocess = (data) => {
        return data.map((item, index) => {
            return {
                ProposalCid: item.ProposalCid["/"],
                DealID: item.DealID,
                Client: item.Proposal.Client,
                Provider: item.Proposal.Provider,
                State: dealState[item.State],
                PieceStatus: item.PieceStatus,
                SectorNumber: item.SectorNumber,
            }
        })
    }

    // preprocess data
    let data = preprocess(deals)

    // init table
    const columns = [
        {
            title: 'Proposal Cid',
            dataIndex: 'ProposalCid',
            render: (_, record) => (<InShort text={record.ProposalCid} />)
        },
        {
            title: 'Deal ID',
            dataIndex: 'DealID',
            sorter: (a, b) => a.DealID - b.DealID,
        },
        {
            title: 'Sector',
            dataIndex: 'SectorNumber',
        },
        {
            title: 'Client',
            dataIndex: 'Client',
            filters: getDefaultFilters(data, record => record.Client),
            onFilter: (value, record) => record.Client === value,
            render: (_, record) => (<InShort text={record.Client} />)
        },
        {
            title: 'Provider',
            filters: getDefaultFilters(data, record => record.Provider),
            onFilter: (value, record) => record.Provider === value,
            dataIndex: 'Provider',
        },
        {
            title: 'Deal State',
            dataIndex: 'State',
            filters: getDefaultFilters(data, record => record.State),
            onFilter: (value, record) => record.State === value,
        },
        {
            title: 'Piece State',
            dataIndex: 'PieceStatus',
            filters: getDefaultFilters(data, record => record.PieceStatus),
            onFilter: (value, record) => record.PieceStatus === value,
        },
        // {
        //     // extra info
        //     render: (_, record) => renderExtra(record)
        // }

    ]

    const footer = () => {
        if (selectedRowKeys.length > 0) {
            return (
                <Row >
                    < Col style={{ textAlign: 'left' }}>
                        <span>
                            has selected {selectedRowKeys.length} items:
                        </span>
                    </Col>
                    <Col offset={1} style={{ textAlign: 'left' }}>
                        <Space >
                            <Button type="link" ><FormOutlined style={{ color: 'darkgreen' }} /> set state </Button>
                        </Space>
                    </Col>
                </Row>
            )
        } else {
            return null
        }
    }

    const rowSelection = {
        selectedRowKeys,
        onChange: (newSelectedRowKeys) => {
            setSelectedRowKeys(newSelectedRowKeys);
        }
    };

    const pagination = {
        hideOnSinglePage: true,
        showSizeChanger: true,
        defaultPageSize: 10,
    }

    const table = (
        <Table
            rowKey={record => record.ProposalCid}
            rowSelection={rowSelection}
            columns={columns}
            dataSource={data}
            pagination={pagination}
            footer={footer}
        ></Table>
    )
    return ret(table)
}


const renderExtra = record => {
    const content = (
        <Descriptions
            bordered
            size="small"
            column={1}
        >
            <Descriptions.Item label="pieceCID">{record.pieceCID}</Descriptions.Item>
            <Descriptions.Item label="size">{record.size}</Descriptions.Item>
            <Descriptions.Item label="pricePerEpoch">{record.pricePerEpoch}</Descriptions.Item>
            <Descriptions.Item label="startEpoch">{record.startEpoch}</Descriptions.Item>
            <Descriptions.Item label="duration">{record.duration}</Descriptions.Item>
            <Descriptions.Item label="dealID">{record.dealID}</Descriptions.Item>
            <Descriptions.Item label="activationEpoch">{record.activationEpoch}</Descriptions.Item>
            <Descriptions.Item label="message">{record.message}</Descriptions.Item>
        </Descriptions>
    )
    return (
        <Popover content={content} overlayStyle={{ maxWidth: '70%' }} title="Deal Info" >
            <InfoCircleOutlined />
        </Popover>
    )
}
