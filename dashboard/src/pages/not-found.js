import { useNavigate } from "react-router-dom"
import { Result, Button } from "antd"

export default function NotFound(props) {
    const navigate = useNavigate()
    return (
        <>
            <Result status={"404"} title="404"
                subTitle="Sorry, the page you visited does not exist."
                extra={<Button type="primary" onClick={() => { navigate('/') }} >Back Home</Button>} />
        </>
    )
}
