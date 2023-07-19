import { Space } from 'antd';
import Summary from '@/component/summary';
import MsgList from '@/component/msg-list';
import SealingThreadList from '@/component/sealing-thread-list';
import DealList from '@/component/deal-list';
import Asset from '../component/asset';
import BlockList from '../component/block-list';

function App() {

  return (
    <Space style={{ width: '100%' }} direction='vertical' size={'large'}>
      <Summary />
      <Asset />
      <MsgList />
      <SealingThreadList />
      <BlockList />
      <DealList />
    </Space>
  );
}

export default App;
