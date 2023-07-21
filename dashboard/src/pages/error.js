import { useRouteError } from "react-router-dom";
import { useNavigate, useLocation } from "react-router-dom"
import { Result, Button } from "antd"

export default function ErrorPage() {
    const navigate = useNavigate()
    let error = useRouteError()
    const location = useLocation();

    if (!error) {
        error = location.state?.error;
        console.warn("transport error:", error);
    } else {
        console.warn("catch error:", error);
    }
    if (!error) {
        return (
            <>
                <Result title={"Unknown Error"}
                    extra={<Button type="primary" onClick={() => { navigate('/') }} >Back Home</Button>} />
            </>
        )
    }

    console.error(error);
    return (
        <>
            <Result status={error.status ? error.status : "error"} title={error.statusText}
                subTitle={error.data ? error.data.error : error.message}
                extra={<Button type="primary" onClick={() => { navigate('/') }} >Back Home</Button>} />
        </>
    )
}
