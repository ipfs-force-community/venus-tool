import { Popover } from "antd";
import Paragraph from "antd/es/typography/Paragraph";

export function getDefaultFilters(list, mapper = (e) => e, toLable = (e) => e) {
    let ret = [];
    if (list.length > 0) {
        ret = list.map(mapper);
        ret = ret.filter((e) => e !== undefined);
        // remove duplicate
        ret = [...new Set(ret)];
        ret = ret.map((e) => {
            return {
                text: toLable(e),
                value: e,
            };
        });
    }
    return ret
}


export function inShort(str, pre = 8, suf = 8, placeholder = "...") {
    if (str.length > pre + suf + 3) {
        return str.slice(0, pre) + placeholder + str.slice(-suf)
    }
    return str
}


export function InShort(props) {
    const { text, pre, suf, placeholder } = props;
    if (!text) {
        return (<Paragraph style={{ marginBottom: 0 }} >{text}</Paragraph>)
    }
    return (
        <Popover content={text}>
            <Paragraph copyable={{ tooltips: false, text: text }} style={{ marginBottom: 0 }} >{inShort(text, pre, suf, placeholder)}</Paragraph>
        </Popover>
    )
}

export function Copyable(props) {
    const { text } = props;
    return (
        <Paragraph copyable={{ tooltips: false, text: text }} style={{ marginBottom: 0 }} >
            {text}
            {props.children}
        </Paragraph>
    )
}

export const Editable = ({ text, onUpdate }) => {
    return (<Paragraph
        editable={{
            onChange: onUpdate,
        }}
        style={{ marginBottom: 0 }}
    >
        {text}
    </Paragraph>)
}
