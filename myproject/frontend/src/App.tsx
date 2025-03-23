import { useState } from 'react';
import { GenerateAndPrintLabel, ParseLogisticsInfo, OpenPDF } from '../wailsjs/go/main/App';
import './App.css';

// 导入UI组件
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from './components/ui/card';
import { Label } from './components/ui/label';
import { Input } from './components/ui/input';
import { Textarea } from './components/ui/textarea';
import { Button } from './components/ui/button';
import { Switch } from './components/ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './components/ui/select';

function App() {
  const [logisticsInfo, setLogisticsInfo] = useState('');
  const [serviceType, setServiceType] = useState('');
  const [phoneNumber, setPhoneNumber] = useState('');
  const [itemNumber, setItemNumber] = useState('');
  const [quantity, setQuantity] = useState(0);
  const [totalItems, setTotalItems] = useState(1);
  const [warehouse, setWarehouse] = useState('建闽店');
  const [shippingCrate, setShippingCrate] = useState('');
  const [autoPrint, setAutoPrint] = useState(false);
  const [message, setMessage] = useState('');
  const [pdfPath, setPdfPath] = useState('');
  const [generating, setGenerating] = useState(false);

  // 解析复制的文本
  const handleParse = async () => {
    try {
      const result = await ParseLogisticsInfo(logisticsInfo);
      
      if (result.serviceType) setServiceType(result.serviceType);
      if (result.phoneNumber) setPhoneNumber(result.phoneNumber);
      if (result.itemNumber) setItemNumber(result.itemNumber);
      if (result.quantity) setQuantity(result.quantity);
      if (result.shippingCrate) setShippingCrate(result.shippingCrate);
      
      setMessage('解析成功！请确认信息并调整需要修改的部分。');
    } catch (error) {
      setMessage(`解析失败: ${error}`);
    }
  };

  // 生成标签
  const handleGenerate = async () => {
    if (!serviceType || !phoneNumber || !itemNumber || quantity <= 0 || !shippingCrate) {
      setMessage('请确保所有必填字段都已填写');
      return;
    }

    setGenerating(true);
    
    try {
      const data = {
        serviceType,
        phoneNumber,
        itemNumber,
        quantity,
        totalItems,
        warehouse,
        shippingCrate,
        currentTime: '' // 后端会自动填充
      };
      
      const result = await GenerateAndPrintLabel(data, autoPrint);
      setMessage(result);
      // 从返回结果中提取PDF文件路径
      const pdfPathMatch = result.match(/标签已生成: (.+\.pdf)/);
      setPdfPath(pdfPathMatch ? pdfPathMatch[1] : '');
    } catch (error) {
      setMessage(`生成失败: ${error}`);
    } finally {
      setGenerating(false);
    }
  };

  // 打开PDF
  const handleOpenPDF = async () => {
    if (pdfPath) {
      try {
        await OpenPDF(pdfPath);
      } catch (error) {
        setMessage(`打开PDF失败: ${error}`);
      }
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col items-center justify-center p-4">
      <Card className="w-full max-w-3xl bg-gray-800 border-gray-700">
        <CardHeader className="bg-gray-700 rounded-t-xl">
          <CardTitle className="text-2xl font-bold text-center">TEMU物流标签生成器</CardTitle>
        </CardHeader>
        <CardContent className="p-6 space-y-6">
          <div className="space-y-4">
            <Label htmlFor="logisticsInfo" className="text-lg">粘贴物流信息</Label>
            <Textarea 
              id="logisticsInfo" 
              value={logisticsInfo} 
              onChange={(e) => setLogisticsInfo(e.target.value)} 
              placeholder="请复制并粘贴物流信息内容..." 
              className="min-h-[150px] bg-gray-700 border-gray-600 text-white"
            />
            <Button onClick={handleParse} className="w-full bg-blue-600 hover:bg-blue-700">
              解析信息
            </Button>
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="serviceType">物流名称</Label>
              <Input 
                id="serviceType" 
                value={serviceType} 
                onChange={(e) => setServiceType(e.target.value)} 
                className="bg-gray-700 border-gray-600"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="phoneNumber">物流单号</Label>
              <Input 
                id="phoneNumber" 
                value={phoneNumber} 
                onChange={(e) => setPhoneNumber(e.target.value)} 
                className="bg-gray-700 border-gray-600"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="itemNumber">货号</Label>
              <Input 
                id="itemNumber" 
                value={itemNumber} 
                onChange={(e) => setItemNumber(e.target.value)} 
                className="bg-gray-700 border-gray-600"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="quantity">商品数量</Label>
              <Input 
                id="quantity" 
                type="number" 
                value={quantity || ''} 
                onChange={(e) => setQuantity(parseInt(e.target.value) || 0)} 
                className="bg-gray-700 border-gray-600"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="totalItems">总件数</Label>
              <Input 
                id="totalItems" 
                type="number" 
                value={totalItems} 
                onChange={(e) => setTotalItems(parseInt(e.target.value) || 1)} 
                className="bg-gray-700 border-gray-600"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="warehouse">店铺名称</Label>
              <Select value={warehouse} onValueChange={setWarehouse}>
                <SelectTrigger className="bg-gray-700 border-gray-600">
                  <SelectValue placeholder="选择店铺" />
                </SelectTrigger>
                <SelectContent className="bg-gray-700 border-gray-600 text-white">
                  <SelectItem value="建闽店">建闽店</SelectItem>
                  <SelectItem value="通洲店">通洲店</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2 col-span-2">
              <Label htmlFor="shippingCrate">收货仓</Label>
              <Input 
                id="shippingCrate" 
                value={shippingCrate} 
                onChange={(e) => setShippingCrate(e.target.value)} 
                className="bg-gray-700 border-gray-600"
              />
            </div>
            
            <div className="col-span-2 flex items-center space-x-2">
              <Switch 
                id="autoPrint" 
                checked={autoPrint} 
                onCheckedChange={setAutoPrint} 
              />
              <Label htmlFor="autoPrint">自动打印标签</Label>
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4 p-6 pt-0">
          <Button 
            onClick={handleGenerate} 
            disabled={generating}
            className="w-full bg-green-600 hover:bg-green-700"
          >
            {generating ? '生成中...' : '生成标签'}
          </Button>
          
          {pdfPath && (
            <Button 
              onClick={handleOpenPDF} 
              variant="outline"
              className="w-full border-blue-500 text-blue-400 hover:bg-blue-900"
            >
              打开PDF
            </Button>
          )}
          
          {message && (
            <div className={`text-center p-3 rounded-md ${message.includes('失败') ? 'bg-red-900/50' : 'bg-green-900/50'}`}>
              {message}
            </div>
          )}
        </CardFooter>
      </Card>
    </div>
  );
}

export default App;