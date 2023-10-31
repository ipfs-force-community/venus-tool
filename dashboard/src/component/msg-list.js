import { Col, Empty, Row, Select, Table, Popover, Modal, Form, InputNumber, Radio, message, Button } from "antd"
import { CheckCircleOutlined, QuestionCircleOutlined, ClockCircleOutlined, CloseCircleOutlined, ExclamationCircleOutlined, DeleteOutlined, PlusSquareOutlined } from '@ant-design/icons';

import { useState } from "react"
import { Fil, msgStateString, ParseFilString } from "../util"
import { getDefaultFilters, InShort } from "./util";
import Card from "./card";
import { MarkBad, SendMsg, useMsgs, useWallets } from "../fetcher";
import Input from "antd/es/input/Input";


export default function MessageList({ pageSize = 10, wallets = [] }) {
    const [selectedRowKeys, setSelectedRowKeys] = useState([]);
    const [wallet, setWallet] = useState("")

    const [sendMsgFrom] = Form.useForm();

    const { data: remoteWallets, isLoading: walletIsLoading } = useWallets()
    if (wallets.length === 0 && remoteWallets.length !== 0) {
        wallets = remoteWallets
    }
    if (wallet === '' && wallets.length > 0) {
        setWallet(wallets[0])
    }
    const { data: msgs, isLoading: msgsIsLoading, mutate: updateMsg } = useMsgs(wallet)


    const markeBadMsgs = (msgIDs) => {
        if (!msgIDs || msgIDs.length === 0) {
            return
        }
        const content = (
            <div>
                <ul>
                    {msgIDs.map((id) => {
                        return (
                            <li key={id}>{id}</li>
                        )
                    })}
                </ul>
            </div>
        )
        Modal.confirm(
            {
                icon: <ExclamationCircleOutlined />,
                okText: 'Yes',
                okType: 'danger',
                cancelText: 'No',
                title: 'Are you sure to mark bad these messages?',
                content: content,
                onOk: (close) => {
                    var promises = msgIDs.map((id) => {
                        let data = {
                            ID: id,
                        }
                        return MarkBad(data)
                    })
                    Promise.all(promises).then((res) => {
                        console.log("mark bad msg success!")
                        message.success("mark bad msg success!")
                        updateMsg()
                        close()
                    }).catch((err) => {
                        console.warn("mark bad msg err:", err)
                        message.error(err.data.error)
                    })
                }
            })
    }

    const onAdd = () => {
        const content = (
            <>
                <Form form={sendMsgFrom} style={{ width: "900" }}>
                    <Form.Item label="From" name="From" initialValue={wallet}>
                        <Select options={wallets.map((wallet) => { return { label: wallet, value: wallet } })} />
                    </Form.Item>
                    <Form.Item label="To" name="To" initialValue={""}>
                        <Input />
                    </Form.Item>
                    <Form.Item label="Value" name="Value" initialValue={0}>
                        <Input />
                    </Form.Item>
                    <Form.Item label="Method" name="Method" initialValue={0}>
                        <InputNumber />
                    </Form.Item>
                    <Form.Item label="Max Fee" name="MaxFee" initialValue={0}>
                        <InputNumber />
                    </Form.Item>
                    <Form.Item label="Gas Over Estimation" name="GasOverEstimation" initialValue={0}>
                        <InputNumber />
                    </Form.Item>
                    <Form.Item label="Gas Over Premium" name="GasOverPremium" initialValue={0}>
                        <InputNumber />
                    </Form.Item>
                    <Form.Item label="Params Type" name="ParamsType" initialValue={"base64"}>
                        <Radio.Group>
                            <Radio value={"base64"}>base64</Radio>
                            <Radio value={"json"}>json</Radio>
                            <Radio value={"hex"}>hex</Radio>
                        </Radio.Group>
                    </Form.Item>
                    <Form.Item label="Params" name="Params" initialValue={""}>
                        <Input.TextArea autoSize allowClear />
                    </Form.Item>
                </Form>
            </>
        )
        Modal.confirm(
            {
                okText: 'Send',
                okType: 'danger',
                cancelText: 'Cancel',
                title: 'Send Message',
                content: content,
                width: 800,
                maskClosable: true,
                onOk: (close) => {
                    let v = sendMsgFrom.getFieldsValue()
                    console.log(v);

                    let data = {
                        From: v.From,
                        To: v.To,
                        Value: ParseFilString(v.Value).toString(),
                        Method: v.Method,
                        SendSpec: {
                            GasOverEstimation: v.GasOverEstimation,
                            GasOverPremium: v.GasOverPremium,
                            MaxFee: v.MaxFee.toString(),
                        }
                    }
                    if (v.Params.length > 0) {
                        data["Params"] = {
                            Data: v.Params,
                            EncType: v.ParamsType,
                        }
                    }

                    console.log(data);

                    SendMsg(data).then((res) => {
                        console.log("send msg res:", res)
                        message.success("send msg success:" + res)
                        updateMsg()
                        close()
                    }).catch((err) => {
                        console.warn("send msg err:", err)
                        message.error(err.data.error)
                    })
                }
            }
        )
    }

    const ret = function (Content) {
        const AddButton = () => {
            return (
                <>
                    <PlusSquareOutlined />
                </>
            )
        }
        return (
            <Card title='Messages' extra={<Button type="link" onClick={onAdd} ><AddButton /> </Button>} >
                <Content />
            </Card>
        )
    }

    if (!msgs) {
        return ret(() => <Empty description='no message'></Empty>)
    }


    // pre set order
    msgs.sort(sortByStateAndTime)

    const footer = (data) => {
        if (selectedRowKeys.length > 0) {
            return (
                <Row >
                    < Col style={{ textAlign: 'left' }}>
                        <span>
                            has selected {selectedRowKeys.length} items:
                        </span>
                    </Col>
                    <Col offset={1} style={{ textAlign: 'left' }}>
                        <Button type="link" onClick={() => markeBadMsgs(selectedRowKeys)}><DeleteOutlined /> mark bad </Button>
                    </Col>
                </Row>
            )
        } else {
            return null
        }
    }


    const WalletSelector = () => {
        const walletOpt = wallets.map((wallet) => {
            return { label: wallet, value: wallet }
        })
        if (wallets.length === 0) {
            return null
        } else if (wallets.length === 1) {
            return (
                <div style={{ textAlign: 'left', paddingBottom: '10px' }}>
                    wallet address: {wallets[0]}
                </div>
            )
        }
        return (
            <div style={{ textAlign: 'left', paddingBottom: '10px' }}>
                wallet address: <Select defaultValue={wallet} options={walletOpt} bordered={false} onChange={setWallet} />
            </div>
        )
    }

    const MsgTable = () => {
        const onSelectChange = (newSelectedRowKeys) => {
            setSelectedRowKeys(newSelectedRowKeys);
        };
        const rowSelection = {
            selectedRowKeys,
            onChange: onSelectChange,
        };

        const pagination = {
            hideOnSinglePage: true,
            showSizeChanger: true,
            defaultPageSize: pageSize,
        }

        const columns = [
            {
                title: 'ID',
                dataIndex: 'ID',
                render: text => (<InShort text={text} />),
            },
            {
                title: 'Method',
                dataIndex: 'Method',
                filters: getDefaultFilters(msgs, (record) => `${record.MethodName} (${record.Method})`),
                onFilter: (value, record) => `${record.MethodName} (${record.Method})` === value,
                render: (_, record) => `${record.MethodName} (${record.Method})`
            },
            {
                title: 'Value',
                dataIndex: 'Value',
                render: (text) => Fil(text),
            },
            {
                title: 'To',
                dataIndex: 'To',
                filters: getDefaultFilters(msgs, (msg) => msg.To),
                onFilter: (value, record) => record.To === value,
                render: text => (<InShort text={text} />),
            },
            {
                title: 'Gas Limit',
                dataIndex: 'GasLimit',
                render: (text) => Fil(text),
            },
            {
                title: "Gas Fee Cap",
                dataIndex: 'GasFeeCap',
                render: (text) => Fil(text),
            }, {
                title: 'Gas Premium',
                dataIndex: 'GasPremium',
                render: (text) => Fil(text),
            },
            {
                title: 'State',
                dataIndex: 'State',
                render: (_, record) => renderState(record),
                sorter: sortByStateAndTime,
                filters: getDefaultFilters(msgs, (msg) => msg.State, msgStateString),
                onFilter: (value, record) => record.State === value,
            },
            {
                title: 'Nonce',
                dataIndex: 'Nonce',
            },
            {
                title: 'Time',
                render: (_, record) => new Date(record.UpdatedAt).toLocaleString(),
                sorter: sortTimeStr,
            },
            {
                key: "operate",
                render: (_, record) => { return (<MarkBadButton enable={record.State === 1} msgid={record.ID} update={updateMsg} />) }
            }
        ]

        return (
            <>
                <WalletSelector />
                <Table rowSelection={rowSelection} loading={msgsIsLoading} columns={columns} rowKey='ID' dataSource={msgs} footer={footer} pagination={pagination} />
            </>
        )
    }
    return ret(MsgTable)
}


const MarkBadButton = (props) => {
    const { enable, msgid, update } = props
    const style = enable ? { color: 'darkred' } : { color: "#a9a9a9" }
    return (
        <DeleteOutlined style={style} onClick={() => {
            if (enable) {
                Modal.confirm(
                    {
                        icon: <ExclamationCircleOutlined />,
                        okText: 'Yes',
                        okType: 'danger',
                        cancelText: 'No',
                        title: 'Mark Bad',
                        content: `Mark message ${msgid} as bad?`,
                        onOk: () => {
                            console.log("mark bad msg:", msgid)
                            let data = {
                                ID: msgid,
                            }
                            MarkBad(data).then((res) => {
                                console.log("mark bad msg success!")
                                message.success("mark bad msg success!")
                                update()
                            }).catch((err) => {
                                console.warn("mark bad msg err:", err)
                                message.error(err.data.error)
                            })
                        }
                    }
                )
            }
        }} />
    )
}



const sortByStateAndTime = function (a, b) {
    const weight = {
        // UnKnown
        0: 5,
        // UnFillMsg
        1: 4,
        // FillMsg
        2: 3,
        // NonceConflictMsg
        5: 2,
        // FailedMsg
        4: 1,
        // onChainMsg
        3: 0,
    }
    if (weight[a.State] < weight[b.State]) {
        return 1;
    }
    if (weight[a.State] > weight[b.State]) {
        return -1;
    }
    // UpdateAt
    const timeA = new Date(a.UpdatedAt)
    const timeB = new Date(b.UpdatedAt)
    if (timeA < timeB) {
        return 1;
    }
    if (timeA > timeB) {
        return -1;
    }
    // ExitCode
    if (a.ExitCode < b.ExitCode) {
        return -1;
    }
    if (a.ExitCode > b.ExitCode) {
        return 1;
    }
    return 0;
}


const sortTimeStr = function (a, b) {
    const timeA = new Date(a.UpdatedAt)
    const timeB = new Date(b.UpdatedAt)
    if (timeA < timeB) {
        return 1;
    }
    if (timeA > timeB) {
        return -1;
    }
    return 0;
}

const renderState = function (record) {
    switch (record.State) {
        case 1:
            // UnFillMsg
            if (record.Receipt.ExitCode === 0) {
                return (<Popover content={'UnFillMsg'} >
                    <ClockCircleOutlined style={{ color: 'darkOrange' }} />
                </Popover>)
            } else {
                return (<Popover overlayStyle={{ width: '50%' }} content={record.ErrorMsg} title="UnFillMsg">
                    <ExclamationCircleOutlined style={{ color: 'darkOrange' }} />
                </Popover>)
            }
        case 2:
            // FillMsg
            return (<Popover content={'FillMsg'} >
                <ClockCircleOutlined style={{ color: 'darkblue' }} />
            </Popover>)

        case 3:
            // OnChainMsg
            if (record.Receipt.ExitCode === 0) {
                // exec success
                return (<Popover content={'OnChainMsg'} >
                    <CheckCircleOutlined style={{ color: 'darkgreen' }} />
                </Popover>)
            } else {
                return (<Popover content={`OnChainMsg (exit:${record.Receipt.ExitCode})`} >
                    <ExclamationCircleOutlined style={{ color: 'darkgreen' }} />
                </Popover>)
            }
        case 4:
            // FailedMsg
            if (record.Receipt.Return === null) {
                return (<Popover content={`FailedMsg (exit:${record.Receipt.ExitCode})`} >
                    <CloseCircleOutlined style={{ color: 'darkred' }} />
                </Popover>)
            } else {
                return (<Popover overlayStyle={{ width: '50%' }} content={record.Receipt.Return} title={`FailedMsg (exit:${record.Receipt.ExitCode})`}>
                    <CloseCircleOutlined style={{ color: 'darkred' }} />
                </Popover>)
            }
        case 5:
            // NonceConflictMsg
            return (<Popover content={'NonceConflictMsg'} >
                <ExclamationCircleOutlined style={{ color: 'darkred' }} />
            </Popover>)
        case 0:
        default:
            // UnKnown
            return (<Popover content='UnKnownState' >
                <QuestionCircleOutlined style={{ color: 'darkgrey' }} />
            </Popover>)
    }
}
