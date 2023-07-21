import './index.css';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { createBrowserRouter, Route, RouterProvider, Routes, createRoutesFromElements } from "react-router-dom";
import Provider from './provider';
import Board from '@/pages/board';
import Layout from '@/pages/layout';
import MsgDetail from '@/pages/msg-detail';
import MinerDetail from '@/pages/miner-detail';
import WalletDetail from './pages/wallet-detail';
import DealDetail from './pages/deal-detail';
import NotFound from './pages/not-found';
import ErrorPage from './pages/error';

const router = createRoutesFromElements(
  <Route path="/" >

    <Route path="/" element={<Layout />} >
      <Route path='/' errorElement={<ErrorPage />}  >
        <Route path="/" element={<Board />} />
        <Route path="/message/:id" element={<MsgDetail />} />
        <Route path="/miner/:id" element={<MinerDetail />} />
        <Route path="/wallet/:id" element={<WalletDetail />} />
        <Route path="/deal/:id" element={<DealDetail />} />
        <Route path="404" element={<NotFound />} />
        <Route path="error" element={<ErrorPage />} />
        <Route path="/message/markbad/:id" action={params => {
          console.log("markbad", params);
        }
        } />
        <Route path="*" element={<NotFound />} />
      </Route>
    </Route>
  </Route>
)

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <Provider>
      <RouterProvider router={createBrowserRouter(router)} />
    </Provider>
  </React.StrictMode>
);
