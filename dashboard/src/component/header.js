import logo from '@/asset/venus-hero-logo.png'
import { Col, Row, Select, Input, Affix } from 'antd';
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search as search } from '@/fetcher';

const Search = Input.Search
const tittleStyle = {
    fontSize: '35px',
    marginInlineStart: '0px',
    marginInlineEnd: '0px',
    fontWeight: 'bold',
    color: '#fce62d'
}

export default function Header(props) {
    const [affixed, setAffixed] = useState(false)

    const navigate = useNavigate()

    const searchStyle = {
        verticalAlign: 'middle',
    }

    if (affixed) {
        searchStyle["backgroundColor"] = '#eaeaee'
        searchStyle["boxShadow"] = 'rgb(50 50 93 / 25%) 0px 6px 12px -2px, rgb(0 0 0 / 30%) 0px 3px 7px -3px'
    }

    const onSearch = value => {

        search(value).then((res) => {
            switch (res.Type) {
                case "miner":
                    navigate(`/miner/${value}`)
                    break
                case "message":
                    navigate(`/message/${value}`)
                    break
                case "deal":
                    navigate(`/deal/${value}`)
                    break
                case "wallet":
                    navigate(`/wallet/${value}`)
                    break
                default:
                    console.error("unknown type", res.Type)
            }

        }).catch((err) => {
            const s = {
                error: {
                    status: err.status,
                    statusText: err.statusText,
                    data: err.data,
                },
            };

            navigate(`/error`, { state: s });
        })
    }

    return (
        <div>
            <Row align="middle" >
                <Col span={2} >
                    <div style={{ height: '64px' }} >
                        <Link to={'/'} >
                            <img src={logo} style={{ height: '64px' }} alt="Venus" />
                        </Link>
                    </div>
                </Col>
                <Col span={2} >
                    <Link to={'/'} ><div style={tittleStyle}>  SSM </div> </Link>
                </Col>
                <Col offset={2} span={12} >
                    <Affix offsetTop={0} onChange={setAffixed}>
                        <Search style={searchStyle} size='middle' className='App-search' placeholder='enter cid or address' width={400} allowClear onSearch={onSearch} />
                    </Affix>
                </Col>
                <Col offset={4} span={2} >
                    <Select size='small' defaultValue={"en"} options={[{ label: 'en', value: 'enUS' }]} />
                </Col>
            </Row>
        </div>
    )
}
