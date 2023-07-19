import { Card as AntdCard } from "antd"
import './card.css'

export default function Card(props) {
    const { title, extra } = props

    const wrap = (
        <div style={{ textAlign: 'left' }}>
            <span className="card-tittle">{title}</span>
        </div>
    )
    return (
        <AntdCard
            title={wrap}
            bordered={false}
            size="large"
            extra={extra}
        >
            {props.children}
        </AntdCard>
    )
}
