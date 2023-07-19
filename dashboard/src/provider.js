import { ConfigProvider } from 'antd';
import { Provider as StoreProvider, useSelector } from 'react-redux';
import { SWRConfig } from 'swr'
import store from './store';
import { DefaultFetcher } from './fetcher';

function InertiaProvider(props) {
    const locale = useSelector(state => state.locale.value);
    return (
        <ConfigProvider locale={locale} componentSize={'small'}>
            {props.children}
        </ConfigProvider>
    )

}

function SWRConfigProvider(props) {
    return (
        <SWRConfig value={{
            // refreshInterval: 10000,
            fetcher: DefaultFetcher
        }}>
            {props.children}
        </SWRConfig>
    )
}

function Provider(props) {
    return (
        <StoreProvider store={store}>
            <InertiaProvider>
                <SWRConfigProvider>
                    {props.children}
                </SWRConfigProvider>
            </InertiaProvider>
        </StoreProvider>
    )
}

export default Provider;
