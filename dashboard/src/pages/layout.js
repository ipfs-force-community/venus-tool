import { Layout as AntLayout, Space } from 'antd';
import AppFooter from '@/component/footer';
import AppHeader from '@/component/header';
import { Outlet } from 'react-router-dom';


const { Header, Footer, Content } = AntLayout;




function Layout() {

    return (
        <div style={{
            verticalAlign: 'middle',
            height: '100vh',
        }}>
            <AntLayout style={{ minHeight: '100%', direction: "ltr" }}>
                <Header className='App-header' style={{ paddingInline: '1.5rem' }} >
                    <AppHeader />
                </Header>
                <Content className='App-content' style={{ backgroundColor: '#f0f2f5', padding: '2.5rem' }} >
                    <Outlet />
                </Content>
                <Footer className='App-footer' >
                    <AppFooter />
                </Footer>
            </AntLayout>
        </div >
    );
}

export default Layout;
