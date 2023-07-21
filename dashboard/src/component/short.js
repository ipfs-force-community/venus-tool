import { Popover } from "antd";
import Paragraph from "antd/es/typography/Paragraph";

export default function Short({ text = "none", pre = 10, suf = 10, placeholder = '...' }) {
    // const { text, pre, suf, placeholder } = props;

    if (text.length > pre + suf + 3) {
        return (
            <Popover content={text}>
                <Paragraph copyable={{ tooltips: false, text: text }} style={{ marginBottom: 0 }} >{text.slice(0, pre) + placeholder + text.slice(-suf)}</Paragraph>
            </Popover>
        )
    }

    return (
        <Paragraph copyable={{ tooltips: false, text: text }} style={{ marginBottom: 0 }} >{text}</Paragraph>
    )
}
