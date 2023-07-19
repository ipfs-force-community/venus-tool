import { Fil, Power } from "@/util"
import { Row, Col, Statistic, Empty } from "antd"
import Card from "./card"
import { useMsgsByUpdate, useWalletInfos, useMinerInfos } from "@/fetcher"
import { Content } from "antd/es/layout/layout"



const shortFil = Fil
const shortPower = Power

export default function Summary(props) {
    let yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    // The seconds should be rounded, avoid duplicate request for every second
    yesterday.setSeconds(0)
    yesterday.setMilliseconds(0)

    const { data: msgs } = useMsgsByUpdate({ updateBefore: yesterday })

    const { data: miners } = useMinerInfos()
    const { data: wallets } = useWalletInfos()


    if (!msgs || !wallets || !miners) {
        return ret(<Empty />)
    }

    const gasUsed = msgs.reduce((acc, msg) => {
        if (msg.Receipt) {
            acc += Number(msg.Receipt.GasUsed)
        }
        return acc
    }, 0)

    let availBalance = wallets.reduce((acc, wallet) => {
        acc += Number(wallet.Balance)
        return acc
    }, 0)
    availBalance += miners.reduce((acc, miner) => {
        acc += Number(miner.AvailBalance)
        acc += Number(miner.MarketBalance.Escrow)
        return acc
    }, 0)

    const totalPower = Number(miners[0] ? miners[0].TotalPower.QualityAdjPower : 0)

    const RBP = miners.reduce((acc, miner) => {
        acc += Number(miner.MinerPower.RawBytePower)
        return acc
    }, 0)
    const QAP = miners.reduce((acc, miner) => {
        acc += Number(miner.MinerPower.QualityAdjPower)
        return acc
    }, 0)

    const minerCollateral = miners.reduce((acc, miner) => {
        acc += Number(miner.LockFunds.InitialPledgeRequirement)
        acc += Number(miner.LockFunds.PreCommitDeposits)
        acc += Number(miner.LockFunds.VestingFunds)
        return acc
    }, 0)
    const marketCollateral = miners.reduce((acc, miner) => {
        acc += Number(miner.MarketBalance.Locked)
        return acc
    }, 0)

    const occupancy = totalPower ? QAP / totalPower : 0
    const winRatePerEpoch = (occupancy * 5) > 1 ? 1 : (occupancy * 5)
    const winRatePerDay = (winRatePerEpoch * 2880).toFixed(0)

    let winCount = 0

    return ret(
        <div style={{ textAlign: 'center' }}>
            <Row gutter={16}>
                <Col span={6}>
                    <Statistic title="Available Balance" value={Fil(availBalance)} />
                </Col>
                <Col span={6}>
                    <Statistic title="Total Collateral" value={Fil(minerCollateral + marketCollateral)} />
                </Col>
                <Col span={6}>
                    <Statistic title="Raw Byte Power" value={Power(RBP)} />
                </Col>
                <Col span={6}>
                    <Statistic title="Quality Adjust Power" value={Power(QAP)} />
                </Col>
            </Row>
            <Row gutter={16}>
                <Col span={6}>
                    <Statistic title="Gas Used" value={`${Fil(gasUsed)}/24h`} />
                </Col>
                <Col span={6}>
                    <Statistic title="Expect Win" value={`${winRatePerDay} block/24h`} />
                </Col>
                <Col span={6}>
                    <Statistic title="Miner Collateral" value={Fil(minerCollateral)} />
                </Col>
                <Col span={6}>
                    <Statistic title="Market Collateral" value={Fil(marketCollateral)} />
                </Col>
            </Row>
        </div>
    )
}

const ret = (content) => {
    return (
        <Card
            title={"Summary"}
        >
            {content}
        </Card>
    )
}
