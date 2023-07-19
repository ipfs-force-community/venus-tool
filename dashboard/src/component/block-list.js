import { Table } from "antd"
import { CheckCircleOutlined, QuestionCircleOutlined, ClockCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { getDefaultFilters, InShort } from "./util";
import Card from "./card";
import { useBlockList, useMiners } from "../fetcher";
import { MinedState } from "../util";

export default function BlockList() {
    const { data: miners } = useMiners()
    const { data: blocks } = useBlockList(miners)

    if (!blocks) {
        return (
            <Card title={"Blocks"} loading={true} />
        )
    }

    const ret = function (content) {
        return (
            <Card title={"Blocks"} >
                {content}
            </Card>
        )
    }



    // preprocess data
    let data = blocks

    // init table
    const columns = [
        {
            title: 'Epoch',
            dataIndex: 'Epoch',
        },
        {
            title: 'Miner',
            dataIndex: 'Miner',
            filters: getDefaultFilters(data, record => record.Miner),
            onFilter: (value, record) => record.Miner === value,
        },
        {
            title: 'MineState',
            filters: getDefaultFilters(data, record => record.MineState),
            onFilter: (value, record) => record.MineState === value,
            dataIndex: 'MineState',
            sort: (a, b) => a.MineState - b.MineState,
            render: (_, r) => renderState(r.MineState)
        },
        {
            title: 'Cid',
            dataIndex: 'Cid',
            render: (text) => (<InShort text={text} />)
        },
        {
            title: 'ParentKey',
            dataIndex: 'ParentKey',
            render: (text) => (<InShort text={text} />)
        },
        {
            title: 'WinningAt',
            dataIndex: 'WinningAt',
            render: (text) => new Date(text).toLocaleString(),
        },
    ]


    const pagination = {
        hideOnSinglePage: true,
        showSizeChanger: true,
        defaultPageSize: 10,
    }

    const table = (
        <Table
            key='Epoch'
            rowKey={record => record.ProposalCid}
            columns={columns}
            dataSource={data}
            pagination={pagination}
        ></Table>
    )
    return ret(table)
}



const renderState = function (state) {
    switch (state) {
        case 0:
            return (<span style={{ color: 'darkblue' }}  > {MinedState(state)} <ClockCircleOutlined /> </span>)
        case 1:
            return (<span style={{ color: 'darkgreen' }}  > {MinedState(state)} <CheckCircleOutlined /> </span>)
        case 5:
            return (<span style={{ color: 'darkgrey' }}   > {MinedState(state)} <QuestionCircleOutlined />  </span>)
        default:
            return (<span style={{ color: 'darkred' }}  > {MinedState(state)}  <CloseCircleOutlined /> </span>)
    }
}
