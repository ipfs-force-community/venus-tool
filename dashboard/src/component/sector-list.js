import { Col, Row, Table, Popover, Space, Descriptions } from "antd"
import { InfoCircleOutlined, FormOutlined } from '@ant-design/icons';
import { useState } from "react"
import Card from "./card";
import { useSectorSum, useSectors } from "../fetcher";
import { Fil } from "../util";

export default function SectorList({ miner }) {
    const { data: sum } = useSectorSum({ miner })
    const [pagination, setPagination] = useState({ hideOnSinglePage: true, current: 1, showSizeChanger: true, pageSize: 100, total: sum })
    const { data: sectors } = useSectors({ miner, pageIndex: pagination.current - 1, pageSize: pagination.total > pagination.pageSize ? pagination.pageSize : pagination.total })

    if (!sectors || !sum) {
        return (
            <Card title="Sectors" loading={true} />
        )
    }
    if (!pagination.total) {
        setPagination({ ...pagination, total: sum })
    }
    const handleTableChange = (pagination) => {
        setPagination(pagination)
    }

    // preprocess data
    let data = sectors

    // init table
    const columns = [
        {
            title: 'ID',
            dataIndex: 'SectorNumber',
        },
        {
            title: 'Activation',
            dataIndex: 'Activation',
        },
        {
            title: 'Expiration',
            dataIndex: 'Expiration',
        },
        {
            title: 'DealWeight',
            dataIndex: 'DealWeight',
        },
        {
            title: 'VerifiedDealWeight',
            dataIndex: 'VerifiedDealWeight',
        },
        {
            title: 'Activation',
            dataIndex: 'Activation',
        }, {
            title: 'InitialPledge',
            dataIndex: 'InitialPledge',
            render: (text) => Fil(text),
        }, {
            title: 'ExpectedDayReward',
            dataIndex: 'ExpectedDayReward',
            render: (text) => Fil(text),
        },
        {
            title: 'Deals',
            dataIndex: 'DealIDs',
        },
    ]


    const table = (
        <Table
            rowKey={record => record.SectorNumber}
            columns={columns}
            dataSource={data}
            pagination={pagination}
            onChange={handleTableChange}
        ></Table>
    )
    return ret(table)
}

const ret = function (content) {
    return (
        <Card title={"Sectors"} >
            {content}
        </Card>
    )
}
