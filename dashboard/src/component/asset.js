import { Tag } from "antd";
import { useNavigate } from "react-router-dom";
import Card from "./card";
import Short from "./short";
import { useWallets, useMiners } from "../fetcher";


export default function Asset() {
    const navigate = useNavigate()
    const { data: wallets, isLoading: walletsIsLoading } = useWallets()
    const { data: miners, isLoading: minersIsLoading } = useMiners()


    const WalletTagss = () => {
        return wallets.map((wallet) => {
            return (
                <Tag color={'gold'} key={wallet} onClick={() => { navigate(`/wallet/${wallet}`) }} style={{ cursor: "pointer" }}>
                    <Short text={wallet} />
                </Tag>
            )
        })
    }

    const MinerTagss = () => miners.map((miner) => {
        return (
            <Tag color={'purple'} key={miner} onClick={() => { navigate(`/miner/${miner}`) }} style={{ cursor: "pointer" }}>
                <Short text={miner} />
            </Tag>
        )
    })


    return (
        <>
            <Card title={"Asset"} loading={walletsIsLoading || minersIsLoading} >
                <WalletTagss />
                <MinerTagss />
            </Card>
        </>
    )
}
