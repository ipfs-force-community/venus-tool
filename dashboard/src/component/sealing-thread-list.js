import { Col, Row, Table, Popover, Space, Descriptions, Modal, Button } from "antd"
import { PauseCircleOutlined, PlayCircleOutlined, ExclamationCircleOutlined, DeleteOutlined } from '@ant-design/icons';
import { useState } from "react"
import { Copyable, getDefaultFilters } from "./util";
import Card from "./card";
import { ApiBase } from "../global";
import axios from "axios";
import { useThreads } from "../fetcher";

export default function SealingThreadList() {
    const [selectedRowKeys, setSelectedRowKeys] = useState([]);
    const { data: threads, mutate: updateThreads } = useThreads()

    const stopThreads = (threads) => {
        if (!threads) {
            return
        }
        var promises = []
        if (Array.isArray(threads)) {
            threads.forEach((thread) => {
                promises.push(axios.put(ApiBase + "/thread/stop", {
                    WorkerName: thread.WorkerName,
                    Index: thread.Index,
                }))
            })
        } else {
            promises.push(axios.put(ApiBase + "/thread/stop", {
                WorkerName: threads.WorkerName,
                Index: threads.Index,
            }))
        }
        Promise.all(promises).then(() => {
            updateThreads()
        })
    }

    const startThreads = (threads) => {
        if (!threads) {
            return
        }
        var promises = []
        if (Array.isArray(threads)) {
            threads.forEach((thread) => {
                promises.push(axios.put(ApiBase + "/thread/start", {
                    WorkerName: thread.WorkerName,
                    Index: thread.Index,
                }))
            })
        } else {
            promises.push(axios.put(ApiBase + "/thread/start", {
                WorkerName: threads.WorkerName,
                Index: threads.Index,
            }))
        }
        Promise.all(promises).then(() => {
            updateThreads()
        })
    }

    const abortThreads = (threads) => {
        if (!threads) {
            return
        }
        var promises = []
        if (Array.isArray(threads)) {
            threads.forEach((thread) => {
                promises.push(axios.put(ApiBase + "/thread/start", {
                    WorkerName: thread.WorkerName,
                    Index: thread.Index,
                }))
            })
        } else {
            promises.push(axios.put(ApiBase + "/thread/start", {
                WorkerName: threads.WorkerName,
                Index: threads.Index,
                State: "Aborted",
            }))
        }
        Promise.all(promises).then(() => {
            updateThreads()
        })
    }

    // prepare data
    let data = preprocess(threads)

    // the footer of sealing thread table
    const footer = () => {
        const listThreads = (threads) => {
            return (
                <div>
                    <ul>
                        {threads.map((thread) => {
                            return (
                                <li key={thread.Key}> {thread.Index}th thread of {thread.WorkerName}</li>
                            )
                        })}
                    </ul>
                </div>
            )
        }

        const onStop = () => {
            const threads = selectedRowKeys.filter((key) => {
                return data.some(record => record.Key === key && !record.Paused)
            }).map((key) => {
                return JSON.parse(key)
            })
            Modal.confirm({
                title: 'Are you sure to stop these threads?',
                icon: <ExclamationCircleOutlined />,
                content: listThreads(threads),
                okText: 'Yes',
                okType: 'danger',
                cancelText: 'No',
                onOk() {
                    stopThreads(threads)
                },
            });
        }
        const onStart = () => {
            const threads = selectedRowKeys.filter((key) => {
                return data.some(record => record.Key === key && record.Paused)
            }).map((key) => {
                return JSON.parse(key)
            })
            Modal.confirm({
                title: 'Are you sure to start these threads?',
                icon: <ExclamationCircleOutlined />,
                content: listThreads(threads),
                okText: 'Yes',
                okType: 'danger',
                cancelText: 'No',
                onOk() {
                    startThreads(threads)
                },
            });
        }

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
                            <Button type="link" onClick={onStart} ><PlayCircleOutlined style={{ color: 'darkgreen' }} /> start  </Button>
                            <Button type="link" onClick={onStop}><PauseCircleOutlined style={{ color: 'darkred' }} /> pause  </Button>
                        </Space>
                    </Col>
                </Row>
            )
        } else {
            return null
        }
    }


    const table = () => {
        const columns = [
            {
                title: 'Worker',
                dataIndex: 'WorkerName',
                filters: getDefaultFilters(data, record => record.WorkerName),
                onFilter: (value, record) => record.WorkerName === value,
                render: (_, record) => renderWorker(record)
            },
            {
                title: 'Index',
                dataIndex: 'Index',
                sorter: (a, b) => a.Index - b.Index,
            },
            {
                title: 'Plan',
                dataIndex: 'Plan',
                filters: getDefaultFilters(data, record => record.Plan),
                onFilter: (value, record) => record.Plan === value,
            },
            {
                title: 'Job ID',
                dataIndex: 'JobID',
            },
            {
                title: 'State',
                dataIndex: 'State',
                filters: getDefaultFilters(data, record => record.State),
                onFilter: (value, record) => record.State === value,
            },
            {
                title: 'Dest',
                dataIndex: 'Dest',
                render: text => (<Copyable text={text} />)
            }, {
                title: 'Location',
                dataIndex: 'Location',
                render: text => (<Copyable >{text}</Copyable>)
            },
            {
                key: 'operate',
                render: (_, record) => renderOperate(record, stopThreads, startThreads, abortThreads),
            }
        ]

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
        return (
            <Table
                rowKey={record => record.Key}
                rowSelection={rowSelection}
                columns={columns}
                dataSource={data}
                pagination={pagination}
                footer={footer}
            ></Table>
        )
    }
    return ret(table)
}

const ret = function (Content) {
    return (
        <div>
            <Card
                title={'Sealing Threads'}
            >
                <Content />
            </Card>
        </div>
    )
}

function preprocess(data) {
    if (!data || !Array.isArray(data)) {
        return []
    }
    // transfer from pre to expect
    return data.map(thread => {
        return {
            Index: thread.index,
            Location: thread.location,
            Plan: thread.plan,
            JobID: thread.job_id,
            Paused: thread.paused,
            PausedElapsed: thread.paused_elapsed,
            State: thread.job_state,
            LastErr: thread.last_error,
            WorkerName: thread.WorkerInfo.Name,
            Dest: thread.WorkerInfo.Dest,
            WorkerInfo: {
                Threads: thread.WorkerInfo.Summary.Threads,
                Empty: thread.WorkerInfo.Summary.Empty,
                Paused: thread.WorkerInfo.Summary.Paused,
                Errors: thread.WorkerInfo.Summary.Errors,
                LastPing: thread.WorkerInfo.LastPing,
            },
            LastPing: thread.LastPing,
            IsLoading: thread.IsLoading ? thread.IsLoading : false,
            Key: JSON.stringify({ WorkerName: thread.WorkerInfo.Name, Index: thread.index })
        }
    })
}

const renderWorker = record => {
    return (
        <Popover
            content={
                <Descriptions
                    title={record.WorkerName}
                    bordered
                    size="small"
                    column={1}
                >
                    <Descriptions.Item label="Dest">{record.WorkerInfo.Dest}</Descriptions.Item>
                    <Descriptions.Item label="Threads">{record.WorkerInfo.Threads}</Descriptions.Item>
                    <Descriptions.Item label="Empty">{record.WorkerInfo.Empty}</Descriptions.Item>
                    <Descriptions.Item label="Paused">{record.WorkerInfo.Paused}</Descriptions.Item>
                    <Descriptions.Item label="Errors">{record.WorkerInfo.Errors}</Descriptions.Item>
                    <Descriptions.Item label="LastPing">{record.LastPing}</Descriptions.Item>
                </Descriptions>
            }
            title="Worker Info"
            trigger="hover"
        >
            {record.WorkerName}
        </Popover>
    )
}

const renderOperate = (record, stopThreads, startThreads, abortThreads) => {
    const stopThread = (workerName, index) => {
        Modal.confirm({
            title: 'Are you sure to stop this thread?',
            icon: <ExclamationCircleOutlined />,
            content: `This action will stop the ${index}th thread of ${workerName}, and it will not be resumed automatically.`,
            okText: 'Yes',
            okType: 'danger',
            cancelText: 'No',
            onOk() {
                stopThreads({ WorkerName: workerName, Index: index })
            }
        })
    }
    const startThread = (workerName, index) => {
        Modal.confirm({
            title: 'Are you sure to start this thread?',
            icon: <ExclamationCircleOutlined />,
            content: `This action will start the ${index}th thread of ${workerName}.`,
            okText: 'Yes',
            okType: 'danger',
            cancelText: 'No',
            onOk() {
                startThreads({ WorkerName: workerName, Index: index })
            }
        })
    }
    const abortThread = (workerName, index) => {
        Modal.confirm({
            title: 'Are you sure to abort this thread?',
            icon: <ExclamationCircleOutlined />,
            content: `This action will abort the ${index}th thread of ${workerName}.`,
            okText: 'Yes',
            okType: 'danger',
            cancelText: 'No',
            onOk() {
                abortThreads({ WorkerName: workerName, Index: index })
            }
        })
    }

    const SpaceWrap = ({ children }) => {
        return (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                <Space>
                    <span />
                    {children}
                    <span />
                </Space>
            </div>
        )
    }

    if (record.Paused) {
        const errInfo = (<>
            <div>
                <Descriptions title='Error Info' bordered column={1} size="small" >
                    <Descriptions.Item label="paused elapsed">{record.PausedElapsed}</Descriptions.Item>
                    <Descriptions.Item label="last error">{record.LastErr}</Descriptions.Item>
                </Descriptions>
            </div>
        </>)
        return (
            <>
                <SpaceWrap>
                    <Popover overlayStyle={{ maxWidth: '80%' }} content={errInfo}>
                        <Button type="link"><PlayCircleOutlined style={{ color: 'darkgreen' }} onClick={() => { startThread(record.WorkerName, record.Index) }} /> </Button>
                    </Popover>
                    <Button type="link"><DeleteOutlined style={{ color: 'darkred' }} onClick={() => { abortThread(record.WorkerName, record.Index) }} /> </Button>
                </SpaceWrap>
            </>

        )
    } else {
        return (
            <SpaceWrap>
                <Button type="link"><PauseCircleOutlined style={{ color: 'darkOrange' }} onClick={() => { stopThread(record.WorkerName, record.Index) }} /> </Button>
            </SpaceWrap>
        )
    }
}
