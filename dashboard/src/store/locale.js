import { createSlice } from '@reduxjs/toolkit';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';

export const localeSlice = createSlice({
    name: 'locale',
    initialState: {
        value: zhCN,
    },
    reducers: {
        setLocale: (state, action) => {
            // Redux Toolkit 允许我们在 reducers 中编写 mutating 逻辑。
            // 它实际上并没有 mutate state 因为它使用了 Immer 库，
            // 它检测到草稿 state 的变化并产生一个全新的基于这些更改的不可变 state
            state.value = action.payload;
        },
    },
});


// 为每个 case reducer 函数生成 Action creators
export const { setLocale } = localeSlice.actions;

export default localeSlice.reducer;
